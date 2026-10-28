using NUnit.Framework;

using io.github.hatayama.UnityCliLoop.Domain;

namespace io.github.hatayama.UnityCliLoop.Tests.Editor
{
    /// <summary>
    /// Test fixture that verifies winget-managed CLI path classification.
    /// </summary>
    public class WingetManagedCliPolicyTests
    {
        /// <summary>
        /// Verifies that WinGet Packages and Links paths are recognized across roots, separators, and casing.
        /// </summary>
        [TestCase(@"C:\Users\<USER_NAME>\AppData\Local\Microsoft\WinGet\Packages\tongguang.tloop_Microsoft.Winget.Source_8wekyb3d8bbwe\tloop.exe")]
        [TestCase(@"C:\Users\<USER_NAME>\AppData\Local\Microsoft\WinGet\Links\tloop.exe")]
        [TestCase(@"C:\Program Files\WinGet\Packages\tongguang.tloop_Microsoft.Winget.Source_8wekyb3d8bbwe\tloop.exe")]
        [TestCase(@"C:\Program Files\WinGet\Links\tloop.exe")]
        [TestCase("C:/Users/<USER_NAME>/AppData/Local/Microsoft/WinGet/Packages/tongguang.tloop_Microsoft.Winget.Source_8wekyb3d8bbwe/tloop.exe")]
        [TestCase(@"C:\Users\<USER_NAME>\AppData\Local\Microsoft\winget\packages\tongguang.tloop\tloop.exe")]
        public void IsWingetManagedPath_WithManagedDirectory_ReturnsTrue(string executablePath)
        {
            bool result = WingetManagedCliPolicy.IsWingetManagedPath(executablePath);

            Assert.That(result, Is.True);
        }

        /// <summary>
        /// Verifies that unrelated paths are not classified as winget-managed.
        /// </summary>
        [TestCase(@"C:\Users\<USER_NAME>\AppData\Local\Programs\tloop\bin\tloop.exe")]
        [TestCase("/opt/homebrew/Cellar/tloop/3.1.0/bin/tloop")]
        [TestCase(@"C:\Tools\WinGet\tloop.exe")]
        [TestCase(@"C:\Tools\WinGet\Other\Packages\tloop.exe")]
        [TestCase(@"D:\Packages\WinGet\tloop.exe")]
        public void IsWingetManagedPath_WithoutAdjacentManagedDirectory_ReturnsFalse(string executablePath)
        {
            bool result = WingetManagedCliPolicy.IsWingetManagedPath(executablePath);

            Assert.That(result, Is.False);
        }

        /// <summary>
        /// Verifies that a missing executable path is not classified as winget-managed.
        /// </summary>
        [Test]
        public void IsWingetManagedPath_WhenPathIsNull_ReturnsFalse()
        {
            bool result = WingetManagedCliPolicy.IsWingetManagedPath(null);

            Assert.That(result, Is.False);
        }
    }
}
