using System;

using UnityEditor;
using UnityEngine;
using UnityEngine.UIElements;

using io.github.hatayama.UnityCliLoop.Domain;
using io.github.hatayama.UnityCliLoop.ToolContracts;

namespace io.github.hatayama.UnityCliLoop.Presentation
{
    /// <summary>
    /// Owns UI Toolkit elements and visual states for the third-party tool migration wizard.
    /// </summary>
    internal sealed class ThirdPartyToolMigrationWizardView
    {
        private const string USS_RELATIVE_PATH = "Editor/Presentation/Setup/SetupWizardWindow.uss";

        private readonly TextField _migrationStatusTextField;
        private readonly ProgressBar _migrationProgressBar;
        private readonly VisualElement _migrationButtonRow;
        private readonly Button _migrateButton;
        private readonly Button _refreshButton;
        private readonly Button _closeButton;

        private ThirdPartyToolMigrationWizardView(
            ScrollView mainScrollView,
            TextField migrationStatusTextField,
            ProgressBar migrationProgressBar,
            VisualElement migrationButtonRow,
            Button migrateButton,
            Button refreshButton,
            Button closeButton)
        {
            Debug.Assert(mainScrollView != null, "mainScrollView must not be null");
            Debug.Assert(migrationStatusTextField != null, "migrationStatusTextField must not be null");
            Debug.Assert(migrationProgressBar != null, "migrationProgressBar must not be null");
            Debug.Assert(migrationButtonRow != null, "migrationButtonRow must not be null");
            Debug.Assert(migrateButton != null, "migrateButton must not be null");
            Debug.Assert(refreshButton != null, "refreshButton must not be null");
            Debug.Assert(closeButton != null, "closeButton must not be null");

            MainScrollView = mainScrollView;
            _migrationStatusTextField = migrationStatusTextField;
            _migrationProgressBar = migrationProgressBar;
            _migrationButtonRow = migrationButtonRow;
            _migrateButton = migrateButton;
            _refreshButton = refreshButton;
            _closeButton = closeButton;
        }

        internal ScrollView MainScrollView { get; }

        internal static ThirdPartyToolMigrationWizardView Create(
            VisualElement root,
            Action refresh,
            Action migrate,
            Action close)
        {
            Debug.Assert(root != null, "root must not be null");
            Debug.Assert(refresh != null, "refresh must not be null");
            Debug.Assert(migrate != null, "migrate must not be null");
            Debug.Assert(close != null, "close must not be null");

            string ussPath = $"{UnityCliLoopConstants.PackageAssetPath}/{USS_RELATIVE_PATH}";
            StyleSheet styleSheet = AssetDatabase.LoadAssetAtPath<StyleSheet>(ussPath);
            Debug.Assert(styleSheet != null, $"USS not found at {ussPath}");
            root.styleSheets.Add(styleSheet);

            ScrollView mainScrollView = CreateMainScrollView(root);
            VisualElement content = CreateCSharpMigrationSection(mainScrollView);
            TextField migrationStatusTextField = CreateStatusTextField(content);
            ProgressBar migrationProgressBar = CreateProgressBar(content);
            VisualElement migrationButtonRow = CreateMigrationButtonRow(content);
            Button migrateButton = CreateMigrateButton(migrationButtonRow);
            (Button refreshButton, Button closeButton) = CreateCSharpMigrationActionSection(mainScrollView);
            CreateFooter(mainScrollView);

            ThirdPartyToolMigrationWizardView view = new ThirdPartyToolMigrationWizardView(
                mainScrollView,
                migrationStatusTextField,
                migrationProgressBar,
                migrationButtonRow,
                migrateButton,
                refreshButton,
                closeButton);
            view.BindEvents(refresh, migrate, close);
            return view;
        }

        internal void ShowNotCheckedState(bool isMigrating)
        {
            _migrationStatusTextField.SetValueWithoutNotify(ThirdPartyToolMigrationWizardText.MigrationNotCheckedText);
            ViewDataBinder.SetVisible(_migrationProgressBar, false);
            ViewDataBinder.SetVisible(_migrationButtonRow, false);
            _migrateButton.SetEnabled(false);
            _migrateButton.text = ThirdPartyToolMigrationWizardText.GetMigrationButtonText(
                isMigrating,
                false,
                false);
            ViewDataBinder.SetVisible(_refreshButton, true);
            ViewDataBinder.SetVisible(_closeButton, false);
            _refreshButton.SetEnabled(true);
        }

        internal void ShowMigrationTargetsState(string[] filePaths, bool isMigrating)
        {
            Debug.Assert(filePaths != null, "filePaths must not be null");

            _migrationStatusTextField.SetValueWithoutNotify(
                ThirdPartyToolMigrationWizardText.GetMigrationStatusText(filePaths.Length));
            ViewDataBinder.SetVisible(_migrationProgressBar, false);
            ViewDataBinder.SetVisible(_migrationButtonRow, true);
            _migrateButton.SetEnabled(!isMigrating);
            _migrateButton.text = ThirdPartyToolMigrationWizardText.GetMigrationButtonText(
                isMigrating,
                true,
                true);
            ViewDataBinder.SetVisible(_refreshButton, true);
            ViewDataBinder.SetVisible(_closeButton, false);
            _refreshButton.SetEnabled(false);
        }

        /// <summary>
        /// Shown right after an auto-scan window open. Mirrors ShowMigrationTargetsState's button
        /// states (Migrate enabled, Check disabled) so only one action is available to click at a time.
        /// </summary>
        internal void ShowAutoScanDetectedState(string[] filePaths, bool isMigrating)
        {
            Debug.Assert(filePaths != null, "filePaths must not be null");

            _migrationStatusTextField.SetValueWithoutNotify(
                ThirdPartyToolMigrationWizardText.GetAutoScanDetectedStatusText(filePaths.Length));
            ViewDataBinder.SetVisible(_migrationProgressBar, false);
            ViewDataBinder.SetVisible(_migrationButtonRow, true);
            _migrateButton.SetEnabled(!isMigrating);
            _migrateButton.text = ThirdPartyToolMigrationWizardText.GetMigrationButtonText(
                isMigrating,
                true,
                true);
            ViewDataBinder.SetVisible(_refreshButton, true);
            ViewDataBinder.SetVisible(_closeButton, false);
            _refreshButton.SetEnabled(false);
        }

        internal void ShowNoMigrationTargetsState(bool isMigrating)
        {
            _migrationStatusTextField.SetValueWithoutNotify(ThirdPartyToolMigrationWizardText.NoMigrationTargetsText);
            ViewDataBinder.SetVisible(_migrationProgressBar, false);
            ViewDataBinder.SetVisible(_migrationButtonRow, true);
            _migrateButton.SetEnabled(false);
            _migrateButton.text = ThirdPartyToolMigrationWizardText.GetMigrationButtonText(
                isMigrating,
                false,
                true);
            ViewDataBinder.SetVisible(_refreshButton, true);
            ViewDataBinder.SetVisible(_closeButton, false);
            _refreshButton.SetEnabled(false);
        }

        /// <summary>
        /// Shown right after a successful Migrate, as distinct wording from ShowNoMigrationTargetsState:
        /// the latter's generic "no migration needed" text reads as an idle/unchanged state, which is
        /// ambiguous with the compile that Migrate's AssetDatabase.Refresh() just kicked off in the
        /// background.
        /// </summary>
        internal void ShowMigrationCompleteState(bool isMigrating)
        {
            _migrationStatusTextField.SetValueWithoutNotify(ThirdPartyToolMigrationWizardText.MigrationCompleteText);
            ViewDataBinder.SetVisible(_migrationProgressBar, false);
            ViewDataBinder.SetVisible(_migrationButtonRow, true);
            _migrateButton.SetEnabled(false);
            _migrateButton.text = ThirdPartyToolMigrationWizardText.GetMigrationButtonText(
                isMigrating,
                false,
                true);
            ViewDataBinder.SetVisible(_refreshButton, true);
            ViewDataBinder.SetVisible(_closeButton, false);
            _refreshButton.SetEnabled(false);
        }

        internal void ShowCheckingState(ThirdPartyToolMigrationProgress progress, bool isMigrating)
        {
            _migrationStatusTextField.SetValueWithoutNotify(
                ThirdPartyToolMigrationWizardText.GetMigrationProgressText(progress, isMigrating));
            ViewDataBinder.SetVisible(_migrationProgressBar, true);
            ViewDataBinder.SetVisible(_migrationButtonRow, true);
            UpdateMigrationProgressBar(progress);
            _migrateButton.SetEnabled(false);
            _migrateButton.text = ThirdPartyToolMigrationWizardText.GetMigrationButtonText(
                isMigrating,
                true,
                true);
            ViewDataBinder.SetVisible(_refreshButton, true);
            ViewDataBinder.SetVisible(_closeButton, false);
            _refreshButton.SetEnabled(false);
        }

        private static ScrollView CreateMainScrollView(VisualElement root)
        {
            ScrollView mainScrollView = new ScrollView();
            mainScrollView.AddToClassList("setup-main-container");
            root.Add(mainScrollView);
            return mainScrollView;
        }

        private static VisualElement CreateCSharpMigrationSection(ScrollView mainScrollView)
        {
            VisualElement migrationSection = new VisualElement();
            migrationSection.AddToClassList("setup-step");
            mainScrollView.Add(migrationSection);

            TextField titleTextField = CreateSelectableText(
                ThirdPartyToolMigrationWizardText.CSharpMigrationSectionTitle,
                "setup-step__title");
            migrationSection.Add(titleTextField);

            VisualElement content = new VisualElement();
            content.AddToClassList("setup-step__content");
            migrationSection.Add(content);

            TextField descriptionTextField = CreateSelectableText(
                ThirdPartyToolMigrationWizardText.CSharpMigrationDescriptionText,
                "setup-step__description-label");
            content.Add(descriptionTextField);
            return content;
        }

        private static TextField CreateStatusTextField(VisualElement content)
        {
            TextField migrationStatusTextField = CreateSelectableText(
                string.Empty,
                "setup-step__status-label");
            migrationStatusTextField.AddToClassList("setup-step__status-label--standalone");
            content.Add(migrationStatusTextField);
            return migrationStatusTextField;
        }

        private static ProgressBar CreateProgressBar(VisualElement content)
        {
            ProgressBar migrationProgressBar = new ProgressBar();
            migrationProgressBar.AddToClassList("setup-progress-bar");
            content.Add(migrationProgressBar);
            return migrationProgressBar;
        }

        private static VisualElement CreateMigrationButtonRow(VisualElement content)
        {
            VisualElement migrationButtonRow = new VisualElement();
            migrationButtonRow.AddToClassList("setup-step__button-row");
            content.Add(migrationButtonRow);
            return migrationButtonRow;
        }

        private static Button CreateMigrateButton(VisualElement migrationButtonRow)
        {
            Debug.Assert(migrationButtonRow != null, "migrationButtonRow must not be null");

            Button migrateButton = new Button();
            migrateButton.text = ThirdPartyToolMigrationWizardText.GetMigrationButtonText(false, false, false);
            migrateButton.AddToClassList("setup-button");
            migrateButton.AddToClassList("setup-button--primary");
            migrateButton.AddToClassList("setup-button--migration-action");
            migrationButtonRow.Add(migrateButton);
            return migrateButton;
        }

        private static (Button refreshButton, Button closeButton) CreateCSharpMigrationActionSection(
            ScrollView mainScrollView)
        {
            VisualElement actionSection = new VisualElement();
            actionSection.AddToClassList("setup-section-actions");
            mainScrollView.Add(actionSection);

            VisualElement buttonRow = new VisualElement();
            buttonRow.AddToClassList("setup-section-actions__button-row");
            actionSection.Add(buttonRow);

            Button refreshButton = new Button();
            refreshButton.text = "Check";
            refreshButton.AddToClassList("setup-button");
            refreshButton.AddToClassList("setup-button--primary");
            buttonRow.Add(refreshButton);

            Button closeButton = new Button();
            closeButton.text = "Close";
            closeButton.AddToClassList("setup-button");
            buttonRow.Add(closeButton);
            return (refreshButton, closeButton);
        }

        private static void CreateFooter(ScrollView mainScrollView)
        {
            VisualElement footer = new VisualElement();
            footer.AddToClassList("setup-footer");
            mainScrollView.Add(footer);

            TextField reopenHintTextField = CreateSelectableText(
                "You can close this wizard and reopen it later from\n" +
                "Window > Tuanjie CLI Loop > Custom Tool Migration.",
                "setup-footer__hint-label");
            footer.Add(reopenHintTextField);
        }

        private static TextField CreateSelectableText(string text, string className)
        {
            TextField textField = new TextField();
            textField.multiline = true;
            textField.isReadOnly = true;
            textField.selectAllOnFocus = false;
            textField.selectAllOnMouseUp = false;
            textField.SetValueWithoutNotify(text);
            textField.AddToClassList("setup-selectable-text");
            textField.AddToClassList(className);
            return textField;
        }

        private void BindEvents(
            Action refresh,
            Action migrate,
            Action close)
        {
            _refreshButton.clicked += refresh;
            _migrateButton.clicked += migrate;
            _closeButton.clicked += close;
        }

        private void UpdateMigrationProgressBar(ThirdPartyToolMigrationProgress progress)
        {
            (int totalItemCount, int processedItemCount) =
                ThirdPartyToolMigrationWizardWindow.GetMigrationProgressBarRange(
                    progress.TotalItemCount,
                    progress.ProcessedItemCount);
            _migrationProgressBar.lowValue = 0;
            _migrationProgressBar.highValue = totalItemCount;
            _migrationProgressBar.value = processedItemCount;
        }
    }
}
