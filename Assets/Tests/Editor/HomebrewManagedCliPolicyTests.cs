using System.Collections.Generic;

using NUnit.Framework;

using io.github.hatayama.UnityCliLoop.Domain;

namespace io.github.hatayama.UnityCliLoop.Tests.Editor
{
    /// <summary>
    /// Test fixture that verifies Homebrew-managed CLI path classification.
    /// </summary>
    public class HomebrewManagedCliPolicyTests
    {
        /// <summary>
        /// Verifies that a Cellar path is classified as Homebrew-managed for every Homebrew prefix.
        /// </summary>
        [TestCase("/opt/homebrew/Cellar/tloop/3.0.0/bin/tloop")]
        [TestCase("/usr/local/Cellar/tloop/3.0.0/bin/tloop")]
        [TestCase("/home/linuxbrew/.linuxbrew/Cellar/tloop/3.0.0/bin/tloop")]
        public void IsHomebrewManagedPath_WithCellarSegment_ReturnsTrue(string executablePath)
        {
            bool result = HomebrewManagedCliPolicy.IsHomebrewManagedPath(
                executablePath,
                directoryPath => false);

            Assert.That(result, Is.True);
        }

        /// <summary>
        /// Verifies that a linked prefix/bin path is Homebrew-managed when its sibling Cellar formula directory exists.
        /// </summary>
        [Test]
        public void IsHomebrewManagedPath_WithLinkedBinPathAndCellarFormulaDirectory_ReturnsTrue()
        {
            List<string> probedDirectories = new();

            bool result = HomebrewManagedCliPolicy.IsHomebrewManagedPath(
                "/opt/homebrew/bin/tloop",
                directoryPath =>
                {
                    probedDirectories.Add(directoryPath);
                    return true;
                });

            Assert.That(result, Is.True);
            Assert.That(probedDirectories, Is.EqualTo(new[] { "/opt/homebrew/Cellar/tloop" }));
        }

        /// <summary>
        /// Verifies that a linked prefix/bin path without a Cellar formula directory is not Homebrew-managed.
        /// </summary>
        [TestCase("/usr/local/bin/tloop")]
        [TestCase("/Users/dev/.local/bin/tloop")]
        public void IsHomebrewManagedPath_WithoutCellarFormulaDirectory_ReturnsFalse(string executablePath)
        {
            bool result = HomebrewManagedCliPolicy.IsHomebrewManagedPath(
                executablePath,
                directoryPath => false);

            Assert.That(result, Is.False);
        }

        /// <summary>
        /// Verifies that an executable outside a prefix/bin directory is never probed as a Homebrew link.
        /// </summary>
        [Test]
        public void IsHomebrewManagedPath_WithNonBinParentDirectory_DoesNotProbeCellar()
        {
            List<string> probedDirectories = new();

            bool result = HomebrewManagedCliPolicy.IsHomebrewManagedPath(
                "/opt/homebrew/sbin/tloop",
                directoryPath =>
                {
                    probedDirectories.Add(directoryPath);
                    return true;
                });

            Assert.That(result, Is.False);
            Assert.That(probedDirectories, Is.Empty);
        }

        /// <summary>
        /// Verifies that a Windows backslash path is normalized instead of being read as a single segment.
        /// </summary>
        [Test]
        public void IsHomebrewManagedPath_WithWindowsBackslashPath_ProbesNormalizedCellarDirectory()
        {
            List<string> probedDirectories = new();

            bool result = HomebrewManagedCliPolicy.IsHomebrewManagedPath(
                "C:\\Users\\dev\\AppData\\Local\\Programs\\tloop\\bin\\tloop.exe",
                directoryPath =>
                {
                    probedDirectories.Add(directoryPath);
                    return false;
                });

            Assert.That(result, Is.False);
            Assert.That(
                probedDirectories,
                Is.EqualTo(new[] { "C:/Users/dev/AppData/Local/Programs/tloop/Cellar/tloop.exe" }));
        }

        /// <summary>
        /// Verifies that a directory name merely containing "Cellar" is not treated as a Cellar segment.
        /// </summary>
        [Test]
        public void IsHomebrewManagedPath_WithPartialCellarSegment_ReturnsFalse()
        {
            bool result = HomebrewManagedCliPolicy.IsHomebrewManagedPath(
                "/Users/dev/CellarTools/tloop",
                directoryPath => false);

            Assert.That(result, Is.False);
        }

        /// <summary>
        /// Verifies that an unknown executable path is never classified as Homebrew-managed.
        /// </summary>
        [TestCase(null)]
        [TestCase("")]
        [TestCase("tloop")]
        public void IsHomebrewManagedPath_WithoutUsablePath_ReturnsFalse(string executablePath)
        {
            bool result = HomebrewManagedCliPolicy.IsHomebrewManagedPath(
                executablePath,
                directoryPath => true);

            Assert.That(result, Is.False);
        }

    }
}
