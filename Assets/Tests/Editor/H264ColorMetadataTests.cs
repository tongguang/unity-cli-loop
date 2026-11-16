using System;
using System.Linq;
using NUnit.Framework;

using io.github.hatayama.UnityCliLoop.FirstPartyTools;

namespace io.github.hatayama.UnityCliLoop.Tests.Editor
{
    public sealed class H264ColorMetadataTests
    {
        // Captured from Unity's Windows H.264 High encoder; includes timing, cropping, and NAL escaping.
        private const string Original = "0164001effe1001c6764001eac2cac15049e7f0110000003001000000303c0da0884538001000468ce3c30";
        // SPS independently produced by FFmpeg h264_metadata with limited range and matrix 6.
        private const string Tagged = "0164001effe1001e6764001eac2cac15049e7f016a04040c80000003008000001e06d044229c01000468ce3c30";

        // Verifies bit-exact agreement with an independent SPS rewrite while retaining the PPS and timing.
        [Test]
        public void PatchConfiguration_MatchesReferenceBitstream()
        {
            byte[] result = H264ColorMetadata.PatchConfiguration(Hex(Original), 6, out byte[] color);

            Assert.That(result, Is.EqualTo(Hex(Tagged)));
            Assert.That(color.Skip(12), Is.EqualTo(new byte[] { 0, 2, 0, 2, 0, 6, 0 }));
        }

        // Verifies an existing explicit SPS matrix is respected rather than replaced by the height fallback.
        [Test]
        public void PatchConfiguration_PreservesExplicitSignalInformation()
        {
            byte[] original = Hex(Tagged);

            byte[] result = H264ColorMetadata.PatchConfiguration(original, 1, out byte[] color);

            Assert.That(result, Is.EqualTo(original));
            Assert.That(color[17], Is.EqualTo(6));
        }

        // Verifies the high-profile extension following the PPS is copied unchanged.
        [Test]
        public void PatchConfiguration_PreservesConfigurationExtension()
        {
            byte[] original = Hex(Original + "fdf8f800");

            byte[] result = H264ColorMetadata.PatchConfiguration(original, 6, out _);

            Assert.That(result, Is.EqualTo(Hex(Tagged + "fdf8f800")));
        }

        // Verifies an SPS without any VUI gains signal information and retains valid RBSP alignment.
        [Test]
        public void PatchConfiguration_CreatesVuiForBaselineSps()
        {
            byte[] original = Hex("0142001effe100076742001ef44b2001000468ce3c30");

            byte[] result = H264ColorMetadata.PatchConfiguration(original, 1, out byte[] color);

            Assert.That(result, Is.EqualTo(Hex("0142001effe1000c6742001ef44b4d408080408001000468ce3c30")));
            Assert.That(color.Skip(12), Is.EqualTo(new byte[] { 0, 2, 0, 2, 0, 1, 0 }));
        }

        // Verifies an existing full-range signal is retained in both the SPS and matching MP4 colr metadata.
        [Test]
        public void PatchConfiguration_PreservesExplicitFullRange()
        {
            // Baseline SPS with video_format=5, full_range=1, and explicit 709 primaries/transfer/matrix.
            byte[] original = Hex("0142001effe1000c6742001ef44b4dc04040408001000468ce3c30");

            byte[] result = H264ColorMetadata.PatchConfiguration(original, 6, out byte[] color);

            Assert.That(result, Is.EqualTo(original));
            Assert.That(color.Skip(12), Is.EqualTo(new byte[] { 0, 1, 0, 1, 0, 1, 128 }));
        }

        // Verifies repeated SPS entries are all updated and the PPS count and data remain intact.
        [Test]
        public void PatchConfiguration_UpdatesEverySps()
        {
            byte[] original = Hex(Original);
            byte[] multiple = original.Take(5).Concat(new byte[] { 226 })
                .Concat(original.Skip(6).Take(30)).Concat(original.Skip(6)).ToArray();
            byte[] expected = Hex(Tagged);
            byte[] expectedMultiple = expected.Take(5).Concat(new byte[] { 226 })
                .Concat(expected.Skip(6).Take(32)).Concat(expected.Skip(6)).ToArray();

            byte[] result = H264ColorMetadata.PatchConfiguration(multiple, 6, out _);

            Assert.That(result, Is.EqualTo(expectedMultiple));
        }

        // Verifies malformed configuration records and SPS data are rejected before rewriting the file.
        [TestCase("")]
        [TestCase("00000000000000")]
        [TestCase("0164001effe000")]
        [TestCase("0164001effe100ff67")]
        [TestCase("0164001effe1000168")]
        [TestCase("0164001effe100046764001e00")]
        public void PatchConfiguration_RejectsMalformedInput(string hex)
        {
            Assert.Throws<System.IO.InvalidDataException>(() => H264ColorMetadata.PatchConfiguration(Hex(hex), 6, out _));
        }

        private static byte[] Hex(string text) => Enumerable.Range(0, text.Length / 2)
            .Select(index => Convert.ToByte(text.Substring(index * 2, 2), 16)).ToArray();
    }
}
