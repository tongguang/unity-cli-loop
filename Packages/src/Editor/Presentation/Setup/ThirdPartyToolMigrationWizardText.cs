using UnityEngine;

using io.github.hatayama.UnityCliLoop.Domain;

namespace io.github.hatayama.UnityCliLoop.Presentation
{
    /// <summary>
    /// Creates user-facing copy for the third-party tool migration wizard.
    /// </summary>
    internal static class ThirdPartyToolMigrationWizardText
    {
        internal const string CSharpMigrationSectionTitle = "C# Source Structure Migration";
        internal const string CSharpMigrationDescriptionText =
            "Use this for C# custom tool source files that still use the V2 API. This section scans the Unity " +
            "project and can rewrite matching C# files automatically.";
        internal const string MigrationNotCheckedText = "C# source migration status has not been checked.";
        internal const string NoMigrationTargetsText = "No C# source structure migration is needed.";
        internal const string MigrationCompleteText =
            "Migration complete. No further C# migration is needed.";
        internal const string MigrationConfirmDialogTitle = "Migrate C# Sources?";
        internal const string MigrationConfirmDialogOkText = "Migrate";
        internal const string MigrationConfirmDialogCancelText = "Cancel";
        private const string MigrationCheckingText = "Scanning C# source files for V3 custom tool API migration...";
        private const string MigrationApplyingText = "Migrating C# source files to V3 custom tool APIs...";
        private const string MigrationButtonReadyText = "Migrate";
        private const string MigrationButtonMigratingText = "Migrating...";
        private const string MigrationButtonNoTargetsText = "Nothing to migrate";
        private const string MigrationButtonCheckRequiredText = "Check required";

        internal static string GetMigrationStatusText(int fileCount)
        {
            Debug.Assert(fileCount >= 0, "fileCount must not be negative");

            string noun = fileCount == 1 ? "file" : "files";
            string verb = fileCount == 1 ? "needs" : "need";
            return $"Found {fileCount} C# {noun} that {verb} V3 migration.";
        }

        /// <summary>
        /// Status shown right after an auto-scan window open: this reflects a compile-error log match,
        /// not a project scan, so the wording says "detected" rather than "found" (Migrate performs the
        /// actual scan). fileCount is 0 for the timeout-fallback path, where HasMigrationTargetsAsync
        /// already confirmed migration is needed but no specific file list is known yet.
        /// </summary>
        internal static string GetAutoScanDetectedStatusText(int fileCount)
        {
            Debug.Assert(fileCount >= 0, "fileCount must not be negative");

            if (fileCount == 0)
            {
                return "Detected legacy V2 custom tool API usage from a compile error. " +
                    "Click Migrate to scan the project and update the affected files.";
            }

            string noun = fileCount == 1 ? "file" : "files";
            return $"Detected {fileCount} C# {noun} using legacy V2 custom tool APIs from a compile error. " +
                "Click Migrate to scan the project and update them.";
        }

        internal static string GetMigrationConfirmDialogMessage(int fileCount)
        {
            Debug.Assert(fileCount >= 0, "fileCount must not be negative");

            if (fileCount == 0)
            {
                return "Files with legacy V2 custom tool API usage will be scanned and rewritten in place.\n\n" +
                    "Commit or back up your project first (VCS recommended).";
            }

            string noun = fileCount == 1 ? "file" : "files";
            return $"{fileCount} {noun} will be rewritten in place.\n\n" +
                "Commit or back up your project first (VCS recommended).";
        }

        internal static string GetMigrationProgressText(
            ThirdPartyToolMigrationProgress progress,
            bool isMigrating)
        {
            string statusText = isMigrating ? MigrationApplyingText : MigrationCheckingText;
            if (progress.TotalItemCount <= 0)
            {
                return statusText;
            }

            return $"{statusText}\n" +
                $"{progress.ProcessedItemCount}/{progress.TotalItemCount} steps complete.";
        }

        internal static string GetMigrationButtonText(
            bool isMigrating,
            bool hasMigrationTargets,
            bool hasCheckedMigrationStatus)
        {
            if (!hasCheckedMigrationStatus)
            {
                return MigrationButtonCheckRequiredText;
            }

            if (isMigrating)
            {
                return MigrationButtonMigratingText;
            }

            return hasMigrationTargets ? MigrationButtonReadyText : MigrationButtonNoTargetsText;
        }
    }
}
