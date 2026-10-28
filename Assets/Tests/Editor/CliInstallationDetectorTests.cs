using System;
using System.Threading;
using System.Threading.Tasks;
using io.github.hatayama.UnityCliLoop.Application;
using UnityEngine;
using NUnit.Framework;

using io.github.hatayama.UnityCliLoop.Infrastructure;

namespace io.github.hatayama.UnityCliLoop.Tests.Editor
{
    /// <summary>
    /// Test fixture that verifies CLI installation detection behavior.
    /// </summary>
    public class CliInstallationDetectorTests
    {
        [Test]
        public void SelectPreferredDetection_WhenShellCommandShadowsPackageOwnedCliUsesShellPath()
        {
            // Verifies that the settings UI reports the same CLI command the user's terminal runs.
            CliInstallationDetection packageOwnedDetection = new(
                "3.0.0-beta.3",
                "/Users/ExampleUser/.local/bin/tloop");
            CliInstallationDetection shellDetection = new(
                "2.1.0",
                "/Users/ExampleUser/.npm-global/bin/tloop");

            CliInstallationDetection result = CliInstallationDetector.SelectPreferredDetection(
                packageOwnedDetection,
                shellDetection);

            Assert.That(result.Version, Is.EqualTo("2.1.0"));
            Assert.That(result.ExecutablePath, Is.EqualTo("/Users/ExampleUser/.npm-global/bin/tloop"));
        }

        [Test]
        public void SelectPreferredDetection_WhenShellCommandMissingUsesPackageOwnedCli()
        {
            // Verifies that package-owned installs still count when the shell cannot resolve uloop.
            CliInstallationDetection packageOwnedDetection = new(
                "3.0.0-beta.3",
                "/Users/ExampleUser/.local/bin/tloop");
            CliInstallationDetection shellDetection = new(
                null,
                null);

            CliInstallationDetection result = CliInstallationDetector.SelectPreferredDetection(
                packageOwnedDetection,
                shellDetection);

            Assert.That(result.Version, Is.EqualTo("3.0.0-beta.3"));
            Assert.That(result.ExecutablePath, Is.EqualTo("/Users/ExampleUser/.local/bin/tloop"));
        }

        [Test]
        public void SelectPreferredDetection_WhenShellCommandExistsButVersionFailsUsesShellPath()
        {
            // Verifies that a broken PATH command is surfaced instead of hidden by the package-owned binary.
            CliInstallationDetection packageOwnedDetection = new(
                "3.0.0-beta.3",
                "/Users/ExampleUser/.local/bin/tloop");
            CliInstallationDetection shellDetection = new(
                null,
                "/Users/ExampleUser/.npm-global/bin/tloop");

            CliInstallationDetection result = CliInstallationDetector.SelectPreferredDetection(
                packageOwnedDetection,
                shellDetection);

            Assert.That(result.Version, Is.Null);
            Assert.That(result.ExecutablePath, Is.EqualTo("/Users/ExampleUser/.npm-global/bin/tloop"));
        }

        [Test]
        public void SelectPreferredDetection_WhenPackageOwnedCliMissingUsesShellPath()
        {
            // Verifies that legacy CLI installs still surface as update candidates.
            CliInstallationDetection packageOwnedDetection = new(
                null,
                "/Users/ExampleUser/.local/bin/tloop");
            CliInstallationDetection shellDetection = new(
                "2.1.0",
                "/Users/ExampleUser/.npm-global/bin/tloop");

            CliInstallationDetection result = CliInstallationDetector.SelectPreferredDetection(
                packageOwnedDetection,
                shellDetection);

            Assert.That(result.Version, Is.EqualTo("2.1.0"));
            Assert.That(result.ExecutablePath, Is.EqualTo("/Users/ExampleUser/.npm-global/bin/tloop"));
        }

        [Test]
        public void SelectPreferredDetection_WhenShellVersionExistsWithoutPathUsesShellVersion()
        {
            // Verifies that installed state does not depend on command path availability.
            CliInstallationDetection packageOwnedDetection = new(
                "3.0.0-beta.3",
                "/Users/ExampleUser/.local/bin/tloop");
            CliInstallationDetection shellDetection = new(
                "2.1.0",
                null);

            CliInstallationDetection result = CliInstallationDetector.SelectPreferredDetection(
                packageOwnedDetection,
                shellDetection);

            Assert.That(result.Version, Is.EqualTo("2.1.0"));
            Assert.That(result.ExecutablePath, Is.Null);
        }

        [Test]
        public void BuildShellCliDetectionCommand_UsesShortVersionFlag()
        {
            // Verifies that shell detection asks the command itself for its terminal-visible version.
            string command = CliShellInstallationProbe.BuildShellCliDetectionCommand("tloop");

            Assert.That(command, Does.Contain("command -v tloop"));
            Assert.That(command, Does.Contain("tloop --version --json"));
            Assert.That(command, Does.Contain("tloop_contract_status=$?"));
            Assert.That(command, Does.Contain("tloop -v"));
            Assert.That(command, Does.Contain("tloop_version_status=$?"));
            Assert.That(command, Does.Contain("__TLOOP_CONTRACT_STATUS_START__"));
            Assert.That(command, Does.Contain("__TLOOP_VERSION_STATUS_START__"));
        }

        [Test]
        public void BuildShellCliDetectionCommandForShell_WhenRuntimeShellIsFish_UsesFishStatusSyntax()
        {
            // Verifies that command syntax follows the actual shell process, not PATH setup support.
            string command = CliShellInstallationProbe.BuildShellCliDetectionCommandForShell(
                "tloop",
                "/opt/homebrew/bin/fish");

            Assert.That(command, Does.Contain("set tloop_version_status $status"));
            Assert.That(command, Does.Contain("set tloop_contract_status $status"));
            Assert.That(command, Does.Not.Contain("tloop_version_status=$?"));
        }

        [Test]
        public void ParseShellCliInstallationOutput_WhenPathAndVersionExist_ReturnsDetection()
        {
            // Verifies that shell detection keeps terminal-visible path data as auxiliary UI context.
            string output = "banner\n"
                            + "__TLOOP_PATH_START__\n"
                            + "/Users/ExampleUser/.npm-global/bin/tloop\n"
                            + "__TLOOP_PATH_END__\n"
                            + "__TLOOP_CONTRACT_START__\n"
                            + "unknown option: --json\n"
                            + "__TLOOP_CONTRACT_END__\n"
                            + "__TLOOP_CONTRACT_STATUS_START__\n"
                            + "1\n"
                            + "__TLOOP_CONTRACT_STATUS_END__\n"
                            + "__TLOOP_VERSION_START__\n"
                            + "2.1.1\n"
                            + "__TLOOP_VERSION_END__\n"
                            + "__TLOOP_VERSION_STATUS_START__\n"
                            + "0\n"
                            + "__TLOOP_VERSION_STATUS_END__\n";

            CliInstallationDetection detection =
                CliShellInstallationProbe.ParseShellCliInstallationOutput(output);

            Assert.That(detection.Version, Is.EqualTo("2.1.1"));
            Assert.That(detection.IsDispatcher, Is.False);
            Assert.That(detection.ExecutablePath, Is.EqualTo("/Users/ExampleUser/.npm-global/bin/tloop"));
        }

        [Test]
        public void ParseShellCliInstallationOutput_WhenLegacyCliJsonExists_ReturnsVersionWithoutProtocol()
        {
            // Verifies pre-dispatcher CLIs cannot satisfy the dispatcher setup contract.
            string output = "__TLOOP_PATH_START__\n"
                            + "/Users/ExampleUser/.npm-global/bin/tloop\n"
                            + "__TLOOP_PATH_END__\n"
                            + "__TLOOP_CONTRACT_START__\n"
                            + "{\"CliVersion\":\"3.0.0-beta.31\",\"ProtocolVersion\":1}\n"
                            + "__TLOOP_CONTRACT_END__\n"
                            + "__TLOOP_CONTRACT_STATUS_START__\n"
                            + "0\n"
                            + "__TLOOP_CONTRACT_STATUS_END__\n"
                            + "__TLOOP_VERSION_START__\n"
                            + "2.1.1\n"
                            + "__TLOOP_VERSION_END__\n"
                            + "__TLOOP_VERSION_STATUS_START__\n"
                            + "0\n"
                            + "__TLOOP_VERSION_STATUS_END__\n";

            CliInstallationDetection detection =
                CliShellInstallationProbe.ParseShellCliInstallationOutput(output);

            Assert.That(detection.Version, Is.EqualTo("3.0.0-beta.31"));
            Assert.That(detection.IsDispatcher, Is.False);
            Assert.That(detection.ExecutablePath, Is.EqualTo("/Users/ExampleUser/.npm-global/bin/tloop"));
        }

        [Test]
        public void ParseShellCliInstallationOutput_WhenDispatcherJsonExists_ReturnsDispatcherDetection()
        {
            // Verifies setup detection recognizes the dispatcher release exposed by global uloop.
            string output = "__TLOOP_PATH_START__\n"
                            + "/Users/ExampleUser/.local/bin/tloop\n"
                            + "__TLOOP_PATH_END__\n"
                            + "__TLOOP_CONTRACT_START__\n"
                            + "{\"DispatcherVersion\":\"3.0.0\"}\n"
                            + "__TLOOP_CONTRACT_END__\n"
                            + "__TLOOP_CONTRACT_STATUS_START__\n"
                            + "0\n"
                            + "__TLOOP_CONTRACT_STATUS_END__\n"
                            + "__TLOOP_VERSION_START__\n"
                            + "3.0.0\n"
                            + "__TLOOP_VERSION_END__\n"
                            + "__TLOOP_VERSION_STATUS_START__\n"
                            + "0\n"
                            + "__TLOOP_VERSION_STATUS_END__\n";

            CliInstallationDetection detection =
                CliShellInstallationProbe.ParseShellCliInstallationOutput(output);

            Assert.That(detection.Version, Is.EqualTo("3.0.0"));
            Assert.That(detection.IsDispatcher, Is.True);
            Assert.That(detection.ExecutablePath, Is.EqualTo("/Users/ExampleUser/.local/bin/tloop"));
        }

        [Test]
        public void IsShellDetectionUsableForPathSetup_WhenLegacyCliVersionIsHigh_ReturnsFalse()
        {
            // Verifies old pre-dispatcher global CLIs cannot satisfy dispatcher path setup.
            CliInstallationDetection detection = new(
                "3.0.0-beta.40",
                "/Users/ExampleUser/.npm-global/bin/tloop");

            bool result = CliShellInstallationProbe.IsShellDetectionUsableForPathSetup(
                detection,
                UnityEngine.RuntimePlatform.OSXEditor,
                (path, platform) => false,
                "3.0.1-beta.6");

            Assert.That(result, Is.False);
        }

        [Test]
        public void ParseShellCliInstallationOutput_WhenProtocolVersionIsOutsideIntRange_ReturnsVersionWithoutProtocol()
        {
            // Verifies oversized protocol metadata cannot break setup compatibility detection.
            string output = "__TLOOP_PATH_START__\n"
                            + "/tmp/tloop\n"
                            + "__TLOOP_PATH_END__\n"
                            + "__TLOOP_CONTRACT_START__\n"
                            + "{\"CliVersion\":\"3.0.0-beta.31\",\"ProtocolVersion\":2147483648}\n"
                            + "__TLOOP_CONTRACT_END__\n"
                            + "__TLOOP_CONTRACT_STATUS_START__\n"
                            + "0\n"
                            + "__TLOOP_CONTRACT_STATUS_END__\n"
                            + "__TLOOP_VERSION_START__\n"
                            + "2.1.1\n"
                            + "__TLOOP_VERSION_END__\n"
                            + "__TLOOP_VERSION_STATUS_START__\n"
                            + "0\n"
                            + "__TLOOP_VERSION_STATUS_END__\n";

            CliInstallationDetection detection =
                CliShellInstallationProbe.ParseShellCliInstallationOutput(output);

            Assert.That(detection.Version, Is.EqualTo("3.0.0-beta.31"));
            Assert.That(detection.ExecutablePath, Is.EqualTo("/tmp/tloop"));
        }

        [Test]
        public void ParseShellCliInstallationOutput_WhenOnlyVersionExists_ReturnsInstalledDetection()
        {
            // Verifies that installation state depends on version output, not path availability.
            string output = "__TLOOP_PATH_START__\n"
                            + "__TLOOP_PATH_END__\n"
                            + "__TLOOP_VERSION_START__\n"
                            + "2.1.1\n"
                            + "__TLOOP_VERSION_END__\n"
                            + "__TLOOP_VERSION_STATUS_START__\n"
                            + "0\n"
                            + "__TLOOP_VERSION_STATUS_END__\n";

            CliInstallationDetection detection =
                CliShellInstallationProbe.ParseShellCliInstallationOutput(output);

            Assert.That(detection.Version, Is.EqualTo("2.1.1"));
            Assert.That(detection.ExecutablePath, Is.Null);
        }

        [Test]
        public void ParseShellCliInstallationOutput_WhenVersionCommandFails_ReturnsPathWithoutVersion()
        {
            // Verifies that failed shell probes do not treat stdout usage text as a CLI version.
            string output = "__TLOOP_PATH_START__\n"
                            + "/Users/ExampleUser/.npm-global/bin/tloop\n"
                            + "__TLOOP_PATH_END__\n"
                            + "__TLOOP_VERSION_START__\n"
                            + "usage: broken tloop\n"
                            + "__TLOOP_VERSION_END__\n"
                            + "__TLOOP_VERSION_STATUS_START__\n"
                            + "1\n"
                            + "__TLOOP_VERSION_STATUS_END__\n";

            CliInstallationDetection detection =
                CliShellInstallationProbe.ParseShellCliInstallationOutput(output);

            Assert.That(detection.Version, Is.Null);
            Assert.That(detection.ExecutablePath, Is.EqualTo("/Users/ExampleUser/.npm-global/bin/tloop"));
        }

        /// <summary>
        /// Verifies that a detector that has not refreshed yet does not report a completed check.
        /// </summary>
        [Test]
        public void IsCheckCompleted_WhenNotRefreshed_ReturnsFalse()
        {
            CliInstallationDetector detector = new CliInstallationDetector(new UnreadablePinReader());

            Assert.That(detector.IsCheckCompleted(), Is.False);
        }

        /// <summary>
        /// Verifies that a detector that has not refreshed yet does not report an installed CLI.
        /// </summary>
        [Test]
        public void IsCliInstalled_WhenNotRefreshed_ReturnsFalse()
        {
            CliInstallationDetector detector = new CliInstallationDetector(new UnreadablePinReader());

            Assert.That(detector.IsCliInstalled(), Is.False);
        }

        /// <summary>
        /// Verifies that on Windows the shell-visibility check succeeds immediately without reading the package pin.
        /// </summary>
        [Test]
        public void IsCliVisibleFromShellAsync_OnWindowsEditor_ReturnsTrueWithoutReadingPin()
        {
            CliInstallationDetector detector = new CliInstallationDetector(new UnreadablePinReader());

            Task<bool> task = detector.IsCliVisibleFromShellAsync(RuntimePlatform.WindowsEditor, CancellationToken.None);

            Assert.That(task.IsCompleted, Is.True);
            Assert.That(task.GetAwaiter().GetResult(), Is.True);
        }

        /// <summary>
        /// Verifies that on POSIX an unreadable package pin fails closed on the caller thread before any shell probe starts.
        /// </summary>
        [Test]
        public void IsCliVisibleFromShellAsync_OnPosixWithUnreadablePin_ThrowsBeforeProbing()
        {
            CliInstallationDetector detector = new CliInstallationDetector(new UnreadablePinReader());

            InvalidOperationException exception = Assert.Throws<InvalidOperationException>(
                () => detector.IsCliVisibleFromShellAsync(RuntimePlatform.OSXEditor, CancellationToken.None));

            Assert.That(exception.Message, Is.EqualTo(UnreadablePinReader.ErrorMessage));
        }

        private sealed class UnreadablePinReader : ICliPinReader
        {
            public const string ErrorMessage = "pin is unreadable in this test";

            public CliPinLoadResult LoadPackagePin()
            {
                throw new InvalidOperationException(ErrorMessage);
            }

            public DispatcherBootstrapPinLoadResult LoadDispatcherBootstrapPin()
            {
                throw new InvalidOperationException(ErrorMessage);
            }

            public string LoadMinimumDispatcherVersionOrThrow()
            {
                throw new InvalidOperationException(ErrorMessage);
            }
        }
    }
}
