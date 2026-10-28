using System;
using System.Collections.Generic;
using System.Threading;
using System.Threading.Tasks;

using UnityEditor;
using UnityEngine;

using io.github.hatayama.UnityCliLoop.Application;
using io.github.hatayama.UnityCliLoop.Domain;
using io.github.hatayama.UnityCliLoop.ToolContracts;

namespace io.github.hatayama.UnityCliLoop.Presentation
{
    /// <summary>
    /// Owns migration-wizard workflow state and async operations for the dedicated Editor window.
    /// </summary>
    internal sealed class ThirdPartyToolMigrationWizardWorkflowController
    {
        private readonly ThirdPartyToolMigrationWizardView _view;
        private readonly ThirdPartyToolMigrationUseCase _thirdPartyToolMigrationUseCase;
        private readonly Action _scheduleResize;
        private readonly IPresentationDialogs _dialogs;
        private readonly IBackgroundWorkRunner _backgroundWorkRunner;

        // Auto-scan seed files (compile-error-matched migration targets) are only used to render the
        // initial detected-state list without scanning; RefreshUI (manual Check / re-check) always
        // does a full, unscoped project scan regardless of these.
        private readonly List<string> _autoScanSeedFilePaths;

        private bool _isMigrating;
        private string[] _pendingMigrationFilePaths = Array.Empty<string>();
        // True once _pendingMigrationFilePaths came from a full-project scan (RefreshUI); false right
        // after an auto-scan detected state, where the count is only a seed estimate that a cascading
        // compile-skip could undercount. The confirm dialog must not assert an exact count in that case.
        private bool _hasVerifiedPendingFileCount = true;
        private CancellationTokenSource _migrationOperationCts;

        internal ThirdPartyToolMigrationWizardWorkflowController(
            ThirdPartyToolMigrationWizardView view,
            ThirdPartyToolMigrationUseCase thirdPartyToolMigrationUseCase,
            List<string> autoScanSeedFilePaths,
            Action scheduleResize,
            IPresentationDialogs dialogs = null,
            IBackgroundWorkRunner backgroundWorkRunner = null)
        {
            Debug.Assert(view != null, "view must not be null");
            Debug.Assert(
                thirdPartyToolMigrationUseCase != null,
                "thirdPartyToolMigrationUseCase must not be null");
            Debug.Assert(autoScanSeedFilePaths != null, "autoScanSeedFilePaths must not be null");
            Debug.Assert(scheduleResize != null, "scheduleResize must not be null");

            _view = view ?? throw new ArgumentNullException(nameof(view));
            _thirdPartyToolMigrationUseCase = thirdPartyToolMigrationUseCase
                ?? throw new ArgumentNullException(nameof(thirdPartyToolMigrationUseCase));
            _autoScanSeedFilePaths = autoScanSeedFilePaths
                ?? throw new ArgumentNullException(nameof(autoScanSeedFilePaths));
            _scheduleResize = scheduleResize
                ?? throw new ArgumentNullException(nameof(scheduleResize));
            _dialogs = dialogs ?? new EditorPresentationDialogs();
            _backgroundWorkRunner = backgroundWorkRunner ?? new ThreadPoolBackgroundWorkRunner();
        }

        internal void ShowInitialState(bool shouldShowAutoScanDetectedState)
        {
            if (shouldShowAutoScanDetectedState)
            {
                ShowAutoScanDetectedState(_autoScanSeedFilePaths);
                return;
            }

            ShowNotCheckedState();
        }

        // seedFilePaths is passed explicitly (rather than reusing the constructor-captured
        // _autoScanSeedFilePaths) because this path re-renders an already-open window: a fresh
        // auto-scan re-detection can find a different file list than the one this controller
        // was originally constructed with.
        internal void TryShowAutoScanDetectedState(bool shouldShowAutoScanDetectedState, List<string> seedFilePaths)
        {
            Debug.Assert(seedFilePaths != null, "seedFilePaths must not be null");

            if (!shouldShowAutoScanDetectedState)
            {
                return;
            }

            ShowAutoScanDetectedState(seedFilePaths);
        }

        internal void ShowAutoScanDetectedState(List<string> seedFilePaths)
        {
            Debug.Assert(seedFilePaths != null, "seedFilePaths must not be null");

            _pendingMigrationFilePaths = seedFilePaths.ToArray();
            _hasVerifiedPendingFileCount = false;
            _view.ShowAutoScanDetectedState(_pendingMigrationFilePaths, _isMigrating);
            _scheduleResize();
        }

        internal async Task RefreshUI()
        {
            CancellationToken ct = BeginMigrationOperation();
            ShowCheckingState(new ThirdPartyToolMigrationProgress(0, 0));
            await Task.Yield();

            string projectRoot = UnityCliLoopPathResolver.GetProjectRoot();
            IProgress<ThirdPartyToolMigrationProgress> progress = CreateProgressReporter(ct);
            ThirdPartyToolMigrationPreview preview;
            try
            {
                preview = await _backgroundWorkRunner.RunTaskAsync(async () =>
                    await _thirdPartyToolMigrationUseCase.PreviewMigrationAsync(projectRoot, progress, ct));
                await MainThreadSwitcher.SwitchToMainThread();
            }
            catch (OperationCanceledException)
            {
                // Cancellation comes from window close or a superseding operation; that owner drives the UI.
                return;
            }
            catch (Exception ex)
            {
                // Without this async-void boundary the exception hits the sync context, the window
                // stays on "Scanning..." forever, and the operation CTS leaks.
                Debug.LogException(ex);
                if (IsMigrationOperationActive(ct))
                {
                    CompleteMigrationOperation(ct);
                    ShowNotCheckedState();
                }

                return;
            }

            if (ct.IsCancellationRequested)
            {
                return;
            }

            CompleteMigrationOperation(ct);
            if (!preview.HasTargets)
            {
                ShowNoMigrationTargetsState();
                return;
            }

            ShowMigrationTargetsState(preview.FilePaths);
        }

        internal async Task HandleMigrateThirdPartyTools()
        {
            int confirmDialogFileCount = ThirdPartyToolMigrationWizardWindow.GetMigrationConfirmDialogFileCount(
                _hasVerifiedPendingFileCount,
                _pendingMigrationFilePaths.Length);
            if (!ThirdPartyToolMigrationWizardWindow.ConfirmMigrationApply(
                confirmDialogFileCount,
                _dialogs.Confirm))
            {
                return;
            }

            CancellationToken ct = BeginMigrationOperation();
            string projectRoot = UnityCliLoopPathResolver.GetProjectRoot();
            ThirdPartyToolMigrationResult result = default;
            bool isMigrationCompletionPending = true;
            _isMigrating = true;
            ShowCheckingState(new ThirdPartyToolMigrationProgress(0, 0));
            await Task.Yield();

            try
            {
                IProgress<ThirdPartyToolMigrationProgress> progress = CreateProgressReporter(ct);
                result = await _backgroundWorkRunner.RunTaskAsync(async () =>
                    await _thirdPartyToolMigrationUseCase.ApplyMigrationAsync(projectRoot, progress, ct));
                if (!ThirdPartyToolMigrationWizardWindow.ShouldFinishMigrationOnMainThread(
                    ct.IsCancellationRequested,
                    result))
                {
                    return;
                }

                await MainThreadSwitcher.SwitchToMainThread();
                if (!ThirdPartyToolMigrationWizardWindow.ShouldFinishMigrationOnMainThread(
                    ct.IsCancellationRequested,
                    result))
                {
                    return;
                }

                CompleteMigrationOperation(ct);
                if (result.Changed)
                {
                    if (!ct.IsCancellationRequested)
                    {
                        ShowCheckingState(new ThirdPartyToolMigrationProgress(1, 1));
                        await Task.Yield();
                    }

                    AssetDatabase.Refresh();
                }

                isMigrationCompletionPending = false;
            }
            catch (OperationCanceledException)
            {
                // The finally block already skips the interrupted-refresh when the token is canceled.
                return;
            }
            catch (Exception ex)
            {
                // PR1 makes a mid-batch apply failure a designed path (rollback, then throw). Log it and
                // return; the finally block sees the still-pending completion and rescans, so the UI
                // reflects the rolled-back files.
                Debug.LogException(ex);
                return;
            }
            finally
            {
                _isMigrating = false;
                bool shouldRefreshAfterInterruptedMigration =
                    ThirdPartyToolMigrationWizardWindow.ShouldRefreshAfterInterruptedMigration(
                        isMigrationCompletionPending,
                        ct.IsCancellationRequested);
                CompleteMigrationOperation(ct);
                if (shouldRefreshAfterInterruptedMigration)
                {
                    await RefreshUI();
                }
            }

            if (ct.IsCancellationRequested)
            {
                return;
            }

            if (ThirdPartyToolMigrationWizardWindow.ShouldRefreshAfterMigration(result))
            {
                await RefreshUI();
                return;
            }

            ShowMigrationCompleteState();
        }

        internal IProgress<ThirdPartyToolMigrationProgress> CreateProgressReporter(CancellationToken ct)
        {
            return new ThirdPartyToolMigrationProgressReporter(
                ct,
                IsMigrationOperationActive,
                ShowCheckingState);
        }

        internal void ShowNotCheckedState()
        {
            _view.ShowNotCheckedState(_isMigrating);
            _scheduleResize();
        }

        internal void ShowMigrationTargetsState(string[] filePaths)
        {
            Debug.Assert(filePaths != null, "filePaths must not be null");

            _pendingMigrationFilePaths = filePaths;
            _hasVerifiedPendingFileCount = true;
            _view.ShowMigrationTargetsState(filePaths, _isMigrating);
            _scheduleResize();
        }

        internal void ShowNoMigrationTargetsState()
        {
            _view.ShowNoMigrationTargetsState(_isMigrating);
            _scheduleResize();
        }

        internal void ShowMigrationCompleteState()
        {
            _view.ShowMigrationCompleteState(_isMigrating);
            _scheduleResize();
        }

        internal void ShowCheckingState(ThirdPartyToolMigrationProgress progress)
        {
            _view.ShowCheckingState(progress, _isMigrating);
        }

        internal CancellationToken BeginMigrationOperation()
        {
            CancelMigrationOperation();
            CancellationTokenSource cts = new CancellationTokenSource();
            _migrationOperationCts = cts;
            return cts.Token;
        }

        internal void CancelMigrationOperation()
        {
            if (_migrationOperationCts == null)
            {
                return;
            }

            _migrationOperationCts.Cancel();
            _migrationOperationCts.Dispose();
            _migrationOperationCts = null;
        }

        internal void CompleteMigrationOperation(CancellationToken ct)
        {
            if (!IsMigrationOperationActive(ct))
            {
                return;
            }

            _migrationOperationCts.Dispose();
            _migrationOperationCts = null;
        }

        internal bool IsMigrationOperationActive(CancellationToken ct)
        {
            return _migrationOperationCts != null && _migrationOperationCts.Token.Equals(ct);
        }
    }
}
