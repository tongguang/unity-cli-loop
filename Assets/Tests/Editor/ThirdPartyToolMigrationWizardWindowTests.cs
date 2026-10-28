using NUnit.Framework;
using UnityEditor;
using UnityEngine;
using UnityEngine.UIElements;

using io.github.hatayama.UnityCliLoop.Domain;
using io.github.hatayama.UnityCliLoop.Presentation;

namespace io.github.hatayama.UnityCliLoop.Tests.Editor
{
    /// <summary>
    /// Test fixture that verifies the dedicated V3 custom tool migration wizard.
    /// </summary>
    public sealed class ThirdPartyToolMigrationWizardWindowTests
    {
        [Test]
        public void ConsumeShouldStartInitialRefresh_WhenPreparedForAutoScan_ReturnsTrueWithoutSessionFlag()
        {
            // Verifies auto-open windows start scanning from the serialized refresh flag alone.
            ThirdPartyToolMigrationWizardWindow window =
                ScriptableObject.CreateInstance<ThirdPartyToolMigrationWizardWindow>();
            try
            {
                ThirdPartyToolMigrationWizardWindow.PrepareForOpen(
                    window,
                    "Tuanjie CLI Loop Migration",
                    new Rect(12f, 34f, 360f, 220f),
                    shouldRefreshAfterCreateGui: true);

                bool shouldStartInitialRefresh = window.ConsumeShouldStartInitialRefresh();

                Assert.That(shouldStartInitialRefresh, Is.True);
            }
            finally
            {
                Object.DestroyImmediate(window);
            }
        }

        [Test]
        public void ConsumeShouldStartInitialRefresh_WhenConsumedTwice_ReturnsFalseOnSecondCall()
        {
            // Verifies the initial-refresh flag is one-shot so CreateGUI cannot double-scan.
            ThirdPartyToolMigrationWizardWindow window =
                ScriptableObject.CreateInstance<ThirdPartyToolMigrationWizardWindow>();
            try
            {
                ThirdPartyToolMigrationWizardWindow.PrepareForOpen(
                    window,
                    "Tuanjie CLI Loop Migration",
                    new Rect(12f, 34f, 360f, 220f),
                    shouldRefreshAfterCreateGui: true);

                bool first = window.ConsumeShouldStartInitialRefresh();
                bool second = window.ConsumeShouldStartInitialRefresh();

                Assert.That(first, Is.True);
                Assert.That(second, Is.False);
            }
            finally
            {
                Object.DestroyImmediate(window);
            }
        }

        [Test]
        public void ConsumeShouldStartInitialRefresh_WhenManualOpen_ReturnsFalse()
        {
            // Verifies manually opened windows wait for an explicit Check instead of auto-scanning.
            ThirdPartyToolMigrationWizardWindow window =
                ScriptableObject.CreateInstance<ThirdPartyToolMigrationWizardWindow>();
            try
            {
                ThirdPartyToolMigrationWizardWindow.PrepareForOpen(
                    window,
                    "Tuanjie CLI Loop Migration",
                    new Rect(12f, 34f, 360f, 220f),
                    shouldRefreshAfterCreateGui: false);

                bool shouldStartInitialRefresh = window.ConsumeShouldStartInitialRefresh();

                Assert.That(shouldStartInitialRefresh, Is.False);
            }
            finally
            {
                Object.DestroyImmediate(window);
            }
        }

        [Test]
        public void GetMigrationStatusText_WhenTargetsExist_ReturnsCountOnlyMessage()
        {
            // Verifies the migration status is a brief count message with no file paths listed.
            string text = ThirdPartyToolMigrationWizardWindow.GetMigrationStatusText(12);

            Assert.That(text, Is.EqualTo("Found 12 C# files that need V3 migration."));
        }

        [Test]
        public void GetMigrationStatusText_WhenSingleTargetExists_UsesSingularNoun()
        {
            // Verifies the singular file count does not pluralize the noun.
            string text = ThirdPartyToolMigrationWizardWindow.GetMigrationStatusText(1);

            Assert.That(text, Is.EqualTo("Found 1 C# file that needs V3 migration."));
        }

        [Test]
        public void ConfirmMigrationApply_WhenDialogIsCanceled_ReturnsFalse()
        {
            // Verifies Cancel leaves migration unstarted by returning false from the confirm helper.
            bool confirmed = ThirdPartyToolMigrationWizardWindow.ConfirmMigrationApply(
                2,
                (title, message, ok, cancel) =>
                {
                    Assert.That(title, Is.EqualTo(ThirdPartyToolMigrationWizardText.MigrationConfirmDialogTitle));
                    Assert.That(message, Does.Contain("2 files will be rewritten in place."));
                    Assert.That(message, Does.Contain("VCS recommended"));
                    Assert.That(ok, Is.EqualTo(ThirdPartyToolMigrationWizardText.MigrationConfirmDialogOkText));
                    Assert.That(cancel, Is.EqualTo(ThirdPartyToolMigrationWizardText.MigrationConfirmDialogCancelText));
                    return false;
                });

            Assert.That(confirmed, Is.False);
        }

        [Test]
        public void ConfirmMigrationApply_WhenDialogIsAccepted_ReturnsTrue()
        {
            // Verifies Migrate confirmation allows the apply path to continue.
            bool confirmed = ThirdPartyToolMigrationWizardWindow.ConfirmMigrationApply(
                1,
                (_, _, _, _) => true);

            Assert.That(confirmed, Is.True);
        }

        [Test]
        public void GetMigrationProgressText_WhenPreviewProgressExists_ReturnsCheckCount()
        {
            // Verifies that the migration wizard reports scan progress while migration targets are unknown.
            ThirdPartyToolMigrationProgress progress = new(3, 12);

            string text = ThirdPartyToolMigrationWizardWindow.GetMigrationProgressText(
                progress,
                isMigrating: false);

            Assert.That(
                text,
                Is.EqualTo("Scanning C# source files for V3 custom tool API migration...\n3/12 steps complete."));
        }

        [Test]
        public void GetMigrationProgressText_WhenApplyProgressExists_ReturnsMigrationCount()
        {
            // Verifies that the migration wizard distinguishes apply progress from preview scans.
            ThirdPartyToolMigrationProgress progress = new(4, 12);

            string text = ThirdPartyToolMigrationWizardWindow.GetMigrationProgressText(
                progress,
                isMigrating: true);

            Assert.That(
                text,
                Is.EqualTo("Migrating C# source files to V3 custom tool APIs...\n4/12 steps complete."));
        }

        [TestCase(false, false, false, "Check required")]
        [TestCase(false, true, true, "Migrate")]
        [TestCase(true, true, true, "Migrating...")]
        [TestCase(false, false, true, "Nothing to migrate")]
        public void GetMigrationButtonText_ReturnsExpectedLabel(
            bool isMigrating,
            bool hasMigrationTargets,
            bool hasCheckedMigrationStatus,
            string expectedLabel)
        {
            // Verifies that the migration action communicates its current state.
            string label = ThirdPartyToolMigrationWizardWindow.GetMigrationButtonText(
                isMigrating,
                hasMigrationTargets,
                hasCheckedMigrationStatus);

            Assert.That(label, Is.EqualTo(expectedLabel));
        }

        [Test]
        public void Create_ShowsOnlyTheCSharpMigrationActions()
        {
            // Verifies that the wizard offers only the C# source migration actions and no AI skill section.
            VisualElement root = new();
            ThirdPartyToolMigrationWizardView.Create(
                root,
                () => { },
                () => { },
                () => { });

            string[] buttonLabels = root.Query<Button>().ToList()
                .ConvertAll(button => button.text)
                .ToArray();

            Assert.That(buttonLabels, Is.EquivalentTo(new[] { "Check required", "Check", "Close" }));
            Assert.That(root.Query<Foldout>().ToList(), Is.Empty);
        }

        [Test]
        public void ShowNoMigrationTargetsState_DisablesCheckButtonWithoutShowingCloseButton()
        {
            // Verifies that the completed migration state keeps Check visible but unavailable.
            VisualElement root = new();
            ThirdPartyToolMigrationWizardView view = ThirdPartyToolMigrationWizardView.Create(
                root,
                () => { },
                () => { },
                () => { });

            view.ShowNoMigrationTargetsState(isMigrating: false);

            Button checkButton = root.Query<Button>().ToList()
                .Find(button => button.text == "Check");
            Button closeButton = root.Query<Button>().ToList()
                .Find(button => button.text == "Close");

            Assert.That(checkButton, Is.Not.Null);
            Assert.That(closeButton, Is.Not.Null);
            Assert.That(checkButton.style.display.value, Is.EqualTo(DisplayStyle.Flex));
            Assert.That(checkButton.enabledSelf, Is.False);
            Assert.That(closeButton.style.display.value, Is.EqualTo(DisplayStyle.None));
        }

        [Test]
        public void ShowMigrationCompleteState_ShowsDistinctWordingFromNoMigrationTargetsState()
        {
            // Verifies that the just-migrated state reads as complete rather than as a generic idle "nothing to migrate", which is ambiguous with the background compile Migrate just triggered.
            VisualElement root = new();
            ThirdPartyToolMigrationWizardView view = ThirdPartyToolMigrationWizardView.Create(
                root,
                () => { },
                () => { },
                () => { });

            view.ShowMigrationCompleteState(isMigrating: false);

            TextField statusTextField = root.Query<TextField>(className: "setup-step__status-label--standalone").First();
            Button checkButton = root.Query<Button>().ToList()
                .Find(button => button.text == "Check");

            Assert.That(statusTextField.value, Is.EqualTo(ThirdPartyToolMigrationWizardText.MigrationCompleteText));
            Assert.That(statusTextField.value, Is.Not.EqualTo(ThirdPartyToolMigrationWizardText.NoMigrationTargetsText));
            Assert.That(checkButton.enabledSelf, Is.False);
        }

        [Test]
        public void ShowAutoScanDetectedState_DisablesCheckButtonWhileMigrateIsAvailable()
        {
            // Verifies that only one action is clickable at a time: Migrate enabled implies Check disabled.
            VisualElement root = new();
            ThirdPartyToolMigrationWizardView view = ThirdPartyToolMigrationWizardView.Create(
                root,
                () => { },
                () => { },
                () => { });

            view.ShowAutoScanDetectedState(new[] { "Assets/Foo.cs" }, isMigrating: false);

            Button migrateButton = root.Query<Button>().ToList()
                .Find(button => button.text == "Migrate");
            Button checkButton = root.Query<Button>().ToList()
                .Find(button => button.text == "Check");

            Assert.That(migrateButton.enabledSelf, Is.True);
            Assert.That(checkButton.enabledSelf, Is.False);
        }

        [TestCase(0, 1, 1, 4, 1000, 100, true)]
        [TestCase(10, 50, 1, 4, 1000, 100, false)]
        [TestCase(10, 110, 1, 4, 1000, 100, true)]
        [TestCase(10, 11, 4, 4, 1000, 100, true)]
        public void ShouldReportMigrationProgress_ReturnsExpectedValue(
            long lastReportTimestamp,
            long currentTimestamp,
            int processedItemCount,
            int totalItemCount,
            long stopwatchFrequency,
            int updateIntervalMilliseconds,
            bool expected)
        {
            // Verifies that migration progress updates are throttled while still reporting completion.
            ThirdPartyToolMigrationProgress progress =
                new(processedItemCount, totalItemCount);

            bool shouldReport = ThirdPartyToolMigrationWizardWindow.ShouldReportMigrationProgress(
                lastReportTimestamp,
                currentTimestamp,
                progress,
                stopwatchFrequency,
                updateIntervalMilliseconds);

            Assert.That(shouldReport, Is.EqualTo(expected));
        }

        [TestCase(false, true, true)]
        [TestCase(true, true, false)]
        [TestCase(false, false, false)]
        public void ShouldApplyMigrationProgress_ReturnsExpectedValue(
            bool isCancellationRequested,
            bool hasActiveOperation,
            bool expected)
        {
            // Verifies that stale migration progress cannot overwrite a rendered preview result.
            bool shouldApply = ThirdPartyToolMigrationWizardWindow.ShouldApplyMigrationProgress(
                isCancellationRequested,
                hasActiveOperation);

            Assert.That(shouldApply, Is.EqualTo(expected));
        }

        [TestCase(1)]
        [TestCase(0)]
        public void ShouldRefreshAfterMigration_WhenMigrationCompletes_ReturnsFalse(int migratedFileCount)
        {
            // Verifies that a completed migration does not immediately start another scan.
            ThirdPartyToolMigrationResult result =
                new(migratedFileCount, migratedFileCount, System.Array.Empty<string>());

            bool shouldRefresh = ThirdPartyToolMigrationWizardWindow.ShouldRefreshAfterMigration(result);

            Assert.That(shouldRefresh, Is.False);
        }

        [TestCase(false, 0, true)]
        [TestCase(true, 0, false)]
        [TestCase(false, 1, true)]
        [TestCase(true, 1, true)]
        public void ShouldFinishMigrationOnMainThread_ReturnsExpectedValue(
            bool isCancellationRequested,
            int migratedFileCount,
            bool expected)
        {
            // Verifies that completed file writes still reach the main-thread asset refresh even after late cancellation.
            ThirdPartyToolMigrationResult result =
                new(migratedFileCount, migratedFileCount, System.Array.Empty<string>());

            bool shouldFinish = ThirdPartyToolMigrationWizardWindow.ShouldFinishMigrationOnMainThread(
                isCancellationRequested,
                result);

            Assert.That(shouldFinish, Is.EqualTo(expected));
        }

        [TestCase(false, 2, 0)]
        [TestCase(true, 2, 2)]
        [TestCase(true, 0, 0)]
        public void GetMigrationConfirmDialogFileCount_ReturnsExpectedValue(
            bool hasVerifiedPendingFileCount,
            int pendingFileCount,
            int expected)
        {
            // Verifies that an unverified seed-derived count never appears in the confirm dialog, since a cascading compile-skip could undercount the real full-scan scope.
            int confirmDialogFileCount = ThirdPartyToolMigrationWizardWindow.GetMigrationConfirmDialogFileCount(
                hasVerifiedPendingFileCount,
                pendingFileCount);

            Assert.That(confirmDialogFileCount, Is.EqualTo(expected));
        }

        [TestCase(0, 5, 1, 0)]
        [TestCase(0, 0, 1, 0)]
        [TestCase(10, 3, 10, 3)]
        [TestCase(10, 15, 10, 10)]
        public void GetMigrationProgressBarRange_ReturnsExpectedValue(
            int totalItemCount,
            int processedItemCount,
            int expectedTotal,
            int expectedProcessed)
        {
            // Verifies that a still-unknown total (0) reports empty progress instead of clamping any inspected count up to a false 100%.
            (int total, int processed) = ThirdPartyToolMigrationWizardWindow.GetMigrationProgressBarRange(
                totalItemCount,
                processedItemCount);

            Assert.That(total, Is.EqualTo(expectedTotal));
            Assert.That(processed, Is.EqualTo(expectedProcessed));
        }

        [TestCase(true, false, true)]
        [TestCase(true, true, false)]
        [TestCase(false, false, false)]
        public void ShouldRefreshAfterInterruptedMigration_ReturnsExpectedValue(
            bool isMigrationCompletionPending,
            bool isCancellationRequested,
            bool expected)
        {
            // Verifies that failed async migrations restore the wizard while user cancellations stay quiet.
            bool shouldRefresh = ThirdPartyToolMigrationWizardWindow.ShouldRefreshAfterInterruptedMigration(
                isMigrationCompletionPending,
                isCancellationRequested);

            Assert.That(shouldRefresh, Is.EqualTo(expected));
        }

        [Test]
        public void PrepareForOpen_PopulatesWindowStateBeforeShowing()
        {
            // Verifies that auto-opened migration windows can request an initial session-gated scan.
            ThirdPartyToolMigrationWizardWindow window =
                ScriptableObject.CreateInstance<ThirdPartyToolMigrationWizardWindow>();
            try
            {
                Rect position = new(12f, 34f, 360f, 220f);

                ThirdPartyToolMigrationWizardWindow.PrepareForOpen(
                    window,
                    "Tuanjie CLI Loop Migration",
                    position,
                    true);

                SerializedObject serializedWindow = new(window);
                SerializedProperty refreshProperty =
                    serializedWindow.FindProperty("_shouldRefreshAfterCreateGui");

                Assert.That(window.titleContent.text, Is.EqualTo("Tuanjie CLI Loop Migration"));
                Assert.That(window.position, Is.EqualTo(position));
                Assert.That(window.minSize, Is.EqualTo(new Vector2(360f, 120f)));
                Assert.That(refreshProperty, Is.Not.Null);
                Assert.That(refreshProperty.boolValue, Is.True);
            }
            finally
            {
                Object.DestroyImmediate(window);
            }
        }

        [Test]
        public void PrepareForOpen_WhenManualOpen_PopulatesWindowStateWithoutInitialRefresh()
        {
            // Verifies that manually opened migration windows wait for the Check button.
            ThirdPartyToolMigrationWizardWindow window =
                ScriptableObject.CreateInstance<ThirdPartyToolMigrationWizardWindow>();
            try
            {
                Rect position = new(12f, 34f, 360f, 220f);

                ThirdPartyToolMigrationWizardWindow.PrepareForOpen(
                    window,
                    "Tuanjie CLI Loop Migration",
                    position,
                    false);

                SerializedObject serializedWindow = new(window);
                SerializedProperty refreshProperty =
                    serializedWindow.FindProperty("_shouldRefreshAfterCreateGui");

                Assert.That(window.titleContent.text, Is.EqualTo("Tuanjie CLI Loop Migration"));
                Assert.That(window.position, Is.EqualTo(position));
                Assert.That(window.minSize, Is.EqualTo(new Vector2(360f, 120f)));
                Assert.That(refreshProperty, Is.Not.Null);
                Assert.That(refreshProperty.boolValue, Is.False);
            }
            finally
            {
                Object.DestroyImmediate(window);
            }
        }

        [Test]
        public void WithContentHeight_UsesMeasuredHeightAndPreservesCenter()
        {
            // Verifies that the migration wizard resizes vertically from measured content height.
            Rect initialRect = new(123f, 456f, 400f, 220f);
            Vector2 frameSize = new(18f, 28f);

            Rect resizedRect =
                ThirdPartyToolMigrationWizardWindow.WithContentHeight(initialRect, 180f, frameSize, 1000f);

            Assert.That(resizedRect.center, Is.EqualTo(initialRect.center));
            Assert.That(resizedRect.size, Is.EqualTo(new Vector2(360f, 208f)));
        }

        [Test]
        public void WithContentHeight_WhenMeasuredHeightIsSmall_ClampsToMinimumHeight()
        {
            // Verifies that content fitting keeps the migration wizard from becoming unusably short.
            Rect initialRect = new(123f, 456f, 400f, 220f);
            Vector2 frameSize = new(18f, 28f);

            Rect resizedRect =
                ThirdPartyToolMigrationWizardWindow.WithContentHeight(initialRect, 12f, frameSize, 1000f);

            Assert.That(resizedRect.center, Is.EqualTo(initialRect.center));
            Assert.That(resizedRect.size, Is.EqualTo(new Vector2(360f, 120f)));
        }

        [Test]
        public void WithContentHeight_WhenCurrentWidthIsWide_UsesSetupWizardWidth()
        {
            // Verifies that content fitting keeps the migration wizard at Setup Wizard width.
            Rect initialRect = new(123f, 456f, 520f, 220f);
            Vector2 frameSize = new(18f, 28f);

            Rect resizedRect =
                ThirdPartyToolMigrationWizardWindow.WithContentHeight(initialRect, 120f, frameSize, 1000f);

            Assert.That(resizedRect.center, Is.EqualTo(initialRect.center));
            Assert.That(resizedRect.size, Is.EqualTo(new Vector2(360f, 148f)));
        }

        [Test]
        public void WithContentHeight_WhenMeasuredHeightExceedsMaxHeight_ClampsToMaxHeight()
        {
            // Verifies that content fitting never grows the migration wizard past the supplied maximum height.
            Rect initialRect = new(123f, 456f, 400f, 220f);
            Vector2 frameSize = new(18f, 28f);

            Rect resizedRect =
                ThirdPartyToolMigrationWizardWindow.WithContentHeight(initialRect, 900f, frameSize, 500f);

            Assert.That(resizedRect.center, Is.EqualTo(initialRect.center));
            Assert.That(resizedRect.size, Is.EqualTo(new Vector2(360f, 500f)));
        }

        [Test]
        public void WithContentHeight_WhenMaxHeightIsBelowMinimum_ClampsToMinimumHeight()
        {
            // Verifies that an unreasonably small maxHeight cannot shrink the window below the usable minimum.
            Rect initialRect = new(123f, 456f, 400f, 220f);
            Vector2 frameSize = new(18f, 28f);

            Rect resizedRect =
                ThirdPartyToolMigrationWizardWindow.WithContentHeight(initialRect, 900f, frameSize, 50f);

            Assert.That(resizedRect.center, Is.EqualTo(initialRect.center));
            Assert.That(resizedRect.size, Is.EqualTo(new Vector2(360f, 120f)));
        }
    }
}
