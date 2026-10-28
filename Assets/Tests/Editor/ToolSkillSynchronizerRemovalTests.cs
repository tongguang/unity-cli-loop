using System;
using System.IO;

using NUnit.Framework;

using io.github.hatayama.UnityCliLoop.Domain;
using io.github.hatayama.UnityCliLoop.Infrastructure;

namespace io.github.hatayama.UnityCliLoop.Tests.Editor
{
    /// <summary>
    /// Test fixture that verifies skill removal against a temporary project root, without touching the running project's files.
    /// </summary>
    public sealed class ToolSkillSynchronizerRemovalTests
    {
        private string _temporaryProjectRoot;

        [SetUp]
        public void SetUp()
        {
            _temporaryProjectRoot = Path.Combine(Path.GetTempPath(), "tloop-test-" + Guid.NewGuid().ToString("N"));
            Directory.CreateDirectory(_temporaryProjectRoot);
        }

        [TearDown]
        public void TearDown()
        {
            if (Directory.Exists(_temporaryProjectRoot))
            {
                Directory.Delete(_temporaryProjectRoot, true);
            }
        }

        /// <summary>
        /// Verifies that removing one tool's skill files leaves installed skills of other tools in both layouts untouched.
        /// </summary>
        [Test]
        public void RemoveSkillFilesAtProjectRoot_WhenInstalledSkillsBelongToOtherTools_KeepsThem()
        {
            string targetRoot = Path.Combine(_temporaryProjectRoot, ".claude");
            string groupedSkillDirectory = Path.Combine(SkillInstallLayout.GetManagedSkillsRoot(targetRoot), "tloop-get-logs");
            string flatSkillDirectory = Path.Combine(SkillInstallLayout.GetSkillsRoot(targetRoot), "tloop-run-tests");
            WriteInstalledSkillFile(groupedSkillDirectory, "---\nname: tloop-get-logs\n---\n");
            WriteInstalledSkillFile(flatSkillDirectory, "---\nname: tloop-run-tests\n---\n");

            ToolSkillSynchronizer.RemoveSkillFilesAtInstallRoot(_temporaryProjectRoot, "compile");

            Assert.That(File.Exists(Path.Combine(groupedSkillDirectory, SkillInstallLayout.SkillFileName)), Is.True);
            Assert.That(File.Exists(Path.Combine(flatSkillDirectory, SkillInstallLayout.SkillFileName)), Is.True);
        }

        private static void WriteInstalledSkillFile(string skillDirectory, string content)
        {
            Directory.CreateDirectory(skillDirectory);
            File.WriteAllText(Path.Combine(skillDirectory, SkillInstallLayout.SkillFileName), content);
        }
    }
}
