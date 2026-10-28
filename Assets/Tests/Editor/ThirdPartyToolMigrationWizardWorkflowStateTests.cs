using System;
using System.Collections.Generic;
using System.Threading;
using System.Threading.Tasks;

using NUnit.Framework;
using UnityEngine.UIElements;

using io.github.hatayama.UnityCliLoop.Application;
using io.github.hatayama.UnityCliLoop.Domain;
using io.github.hatayama.UnityCliLoop.Presentation;

namespace io.github.hatayama.UnityCliLoop.Tests.Editor
{
    /// <summary>
    /// Verifies the migration wizard workflow controller's synchronous state transitions and the
    /// superseding of in-flight operation tokens.
    /// </summary>
    public sealed class ThirdPartyToolMigrationWizardWorkflowStateTests
    {
        private VisualElement _root;
        private int _resizeCount;
        private ThirdPartyToolMigrationWizardWorkflowController _controller;

        [SetUp]
        public void SetUp()
        {
            _root = new VisualElement();
            _resizeCount = 0;
            ThirdPartyToolMigrationWizardView view = ThirdPartyToolMigrationWizardView.Create(
                _root,
                () => { },
                () => { },
                () => { });
            _controller = new ThirdPartyToolMigrationWizardWorkflowController(
                view,
                new ThirdPartyToolMigrationUseCase(new UnusedMigrationPort()),
                new List<string> { "Assets/Seed.cs" },
                () => _resizeCount++,
                new RecordingPresentationDialogs());
        }

        [TearDown]
        public void TearDown()
        {
            _controller.CancelMigrationOperation();
        }

        /// <summary>
        /// Verifies a plain window open shows the not-checked status instead of the seed list.
        /// </summary>
        [Test]
        public void ShowInitialState_WithoutAutoScan_ShowsTheNotCheckedStatus()
        {
            _controller.ShowInitialState(shouldShowAutoScanDetectedState: false);

            Assert.That(GetStatusText(), Is.EqualTo(ThirdPartyToolMigrationWizardText.MigrationNotCheckedText));
            Assert.That(_resizeCount, Is.EqualTo(1));
        }

        /// <summary>
        /// Verifies an auto-scan window open renders the seed files captured at construction.
        /// </summary>
        [Test]
        public void ShowInitialState_WithAutoScan_ShowsTheConstructorSeedCount()
        {
            _controller.ShowInitialState(shouldShowAutoScanDetectedState: true);

            Assert.That(
                GetStatusText(),
                Is.EqualTo(ThirdPartyToolMigrationWizardText.GetAutoScanDetectedStatusText(1)));
            Assert.That(_resizeCount, Is.EqualTo(1));
        }

        /// <summary>
        /// Verifies a re-detection that should not show the detected state keeps the window as it was.
        /// </summary>
        [Test]
        public void TryShowAutoScanDetectedState_WhenNotRequested_LeavesTheCurrentStatus()
        {
            _controller.ShowNoMigrationTargetsState();

            _controller.TryShowAutoScanDetectedState(
                shouldShowAutoScanDetectedState: false,
                new List<string> { "Assets/A.cs", "Assets/B.cs" });

            Assert.That(GetStatusText(), Is.EqualTo(ThirdPartyToolMigrationWizardText.NoMigrationTargetsText));
            Assert.That(_resizeCount, Is.EqualTo(1));
        }

        /// <summary>
        /// Verifies a scan result renders its file count and schedules a resize.
        /// </summary>
        [Test]
        public void ShowMigrationTargetsState_ShowsTheScannedFileCount()
        {
            _controller.ShowMigrationTargetsState(new[] { "Assets/A.cs", "Assets/B.cs", "Assets/C.cs" });

            Assert.That(GetStatusText(), Is.EqualTo(ThirdPartyToolMigrationWizardText.GetMigrationStatusText(3)));
            Assert.That(_resizeCount, Is.EqualTo(1));
        }

        /// <summary>
        /// Verifies a finished migration shows the completion wording and schedules a resize.
        /// </summary>
        [Test]
        public void ShowMigrationCompleteState_ShowsTheCompletionStatus()
        {
            _controller.ShowMigrationCompleteState();

            Assert.That(GetStatusText(), Is.EqualTo(ThirdPartyToolMigrationWizardText.MigrationCompleteText));
            Assert.That(_resizeCount, Is.EqualTo(1));
        }

        /// <summary>
        /// Verifies progress updates render the scanning text but do not reflow the window.
        /// </summary>
        [Test]
        public void ShowCheckingState_ShowsTheScanProgressWithoutResizing()
        {
            ThirdPartyToolMigrationProgress progress = new ThirdPartyToolMigrationProgress(2, 5);

            _controller.ShowCheckingState(progress);

            Assert.That(
                GetStatusText(),
                Is.EqualTo(ThirdPartyToolMigrationWizardText.GetMigrationProgressText(progress, false)));
            Assert.That(_resizeCount, Is.EqualTo(0));
        }

        /// <summary>
        /// Verifies a new migration operation cancels the previous one, which then can no longer complete the
        /// newer operation.
        /// </summary>
        [Test]
        public void BeginMigrationOperation_CancelsAndDeactivatesThePreviousToken()
        {
            CancellationToken first = _controller.BeginMigrationOperation();
            CancellationToken second = _controller.BeginMigrationOperation();

            _controller.CompleteMigrationOperation(first);

            Assert.That(first.IsCancellationRequested, Is.True);
            Assert.That(_controller.IsMigrationOperationActive(first), Is.False);
            Assert.That(_controller.IsMigrationOperationActive(second), Is.True);
        }

        /// <summary>
        /// Verifies completing the active migration operation clears it without signaling cancellation.
        /// </summary>
        [Test]
        public void CompleteMigrationOperation_ForTheActiveToken_EndsItWithoutCanceling()
        {
            CancellationToken token = _controller.BeginMigrationOperation();

            _controller.CompleteMigrationOperation(token);

            Assert.That(_controller.IsMigrationOperationActive(token), Is.False);
            Assert.That(token.IsCancellationRequested, Is.False);
        }

        /// <summary>
        /// Verifies the controller's progress reporter renders progress for the active scan and drops it once that
        /// scan has ended, even though its token was never canceled.
        /// </summary>
        [Test]
        public void CreateProgressReporter_AppliesProgressOnlyWhileTheOperationIsActive()
        {
            CancellationToken stale = _controller.BeginMigrationOperation();
            IProgress<ThirdPartyToolMigrationProgress> staleReporter = _controller.CreateProgressReporter(stale);
            _controller.CompleteMigrationOperation(stale);
            CancellationToken active = _controller.BeginMigrationOperation();
            IProgress<ThirdPartyToolMigrationProgress> activeReporter = _controller.CreateProgressReporter(active);
            _controller.ShowNotCheckedState();

            staleReporter.Report(new ThirdPartyToolMigrationProgress(1, 4));
            string afterStaleReport = GetStatusText();
            activeReporter.Report(new ThirdPartyToolMigrationProgress(3, 4));

            Assert.That(afterStaleReport, Is.EqualTo(ThirdPartyToolMigrationWizardText.MigrationNotCheckedText));
            Assert.That(
                GetStatusText(),
                Is.EqualTo(ThirdPartyToolMigrationWizardText.GetMigrationProgressText(
                    new ThirdPartyToolMigrationProgress(3, 4),
                    false)));
        }

        private string GetStatusText()
        {
            return _root.Query<TextField>(className: "setup-step__status-label--standalone").First().value;
        }

        private sealed class UnusedMigrationPort : IThirdPartyToolMigrationPort
        {
            public ThirdPartyToolMigrationPreview PreviewMigration(string projectRoot)
            {
                throw new NotSupportedException();
            }

            public Task<ThirdPartyToolMigrationPreview> PreviewMigrationAsync(
                string projectRoot,
                IProgress<ThirdPartyToolMigrationProgress> progress,
                CancellationToken ct)
            {
                throw new NotSupportedException();
            }

            public (bool Found, List<string> TargetFilePaths) TryDetectAutoScanTargetsFromCompileErrors(
                string projectRoot)
            {
                throw new NotSupportedException();
            }

            public Task<bool> HasMigrationTargetsAsync(string projectRoot, CancellationToken ct)
            {
                throw new NotSupportedException();
            }

            public ThirdPartyToolMigrationResult ApplyMigration(string projectRoot)
            {
                throw new NotSupportedException();
            }

            public Task<ThirdPartyToolMigrationResult> ApplyMigrationAsync(
                string projectRoot,
                IProgress<ThirdPartyToolMigrationProgress> progress,
                CancellationToken ct)
            {
                throw new NotSupportedException();
            }
        }
    }
}
