using System;
using System.IO;
using System.Linq;
using System.Text;
using NUnit.Framework;

using io.github.hatayama.UnityCliLoop.FirstPartyTools;

namespace io.github.hatayama.UnityCliLoop.Tests.Editor
{
    public sealed class Mp4ColorMetadataTests
    {
        private string _path;

        [SetUp]
        public void SetUp()
        {
            _path = Path.Combine(Path.GetTempPath(), "tloop-colour-" + Guid.NewGuid().ToString("N") + ".mp4");
        }

        [TearDown]
        public void TearDown()
        {
            File.Delete(_path);
        }

        // Verifies the height-only boundary and limited range without inventing primaries or transfer metadata.
        [TestCase(750, 1334, 1)]
        [TestCase(336, 600, 1)]
        [TestCase(324, 578, 1)]
        [TestCase(324, 576, 6)]
        [TestCase(300, 532, 6)]
        [TestCase(186, 332, 6)]
        [TestCase(1280, 480, 6)]
        public void Apply_WritesMeasuredMatrixAndLimitedRange(int width, int height, int matrix)
        {
            byte[] movie = Movie(width, height);
            File.WriteAllBytes(_path, Join(Box("mdat", new byte[] { 31, 42, 53 }), movie));

            Mp4ColorMetadata.Apply(_path, height);

            byte[] result = File.ReadAllBytes(_path);
            byte[] updatedMovie = result.Skip(11 + movie.Length).ToArray();
            byte[] video = Descend(updatedMovie, "trak", "mdia", "minf", "stbl", "stsd", "avc1");
            byte[] color = Children(video, 86).Single(box => Type(box) == "colr");
            Assert.That(color, Is.EqualTo(new byte[]
            {
                0, 0, 0, 19, 99, 111, 108, 114, 110, 99, 108, 120,
                0, 2, 0, 2, 0, (byte)matrix, 0
            }));
            Assert.That(ReadSize(updatedMovie), Is.GreaterThan(movie.Length + 19));
            Assert.That(Children(video, 86).Single(box => Type(box) == "avcC").Length, Is.GreaterThan(47));
        }

        // Verifies that both moov layouts leave encoded bytes, chunk offsets, and unrelated boxes intact.
        [TestCase(false)]
        [TestCase(true)]
        public void Apply_PreservesMediaDataAndChunkOffsets(bool fastStart)
        {
            byte[] movie = Movie(324, 576);
            byte[] media = Box("mdat", new byte[] { 0, 0, 0, 4, 101, 23, 45, 67 });
            byte[] original = fastStart ? Join(movie, media) : Join(media, movie);
            File.WriteAllBytes(_path, original);

            Mp4ColorMetadata.Apply(_path, 576);

            byte[] result = File.ReadAllBytes(_path);
            int movieOffset = fastStart ? 0 : media.Length;
            byte[] retainedPrefix = result.Take(original.Length).ToArray();
            Array.Copy(Encoding.ASCII.GetBytes("moov"), 0, retainedPrefix, movieOffset + 4, 4);
            Assert.That(retainedPrefix, Is.EqualTo(original));
            byte[] updated = result.Skip(original.Length).ToArray();
            byte[] table = Descend(updated, "trak", "mdia", "minf", "stbl");
            Assert.That(Children(table, 8).Single(box => Type(box) == "stco"),
                Is.EqualTo(Box("stco", new byte[] { 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 8 })));
            Assert.That(Children(updated, 8).Last(), Is.EqualTo(Box("udta", new byte[] { 93, 94 })));
        }

        // Verifies repeated application does not append duplicate moov or colr boxes.
        [Test]
        public void Apply_IsIdempotent()
        {
            File.WriteAllBytes(_path, Movie(324, 576));
            Mp4ColorMetadata.Apply(_path, 576);
            byte[] once = File.ReadAllBytes(_path);

            Mp4ColorMetadata.Apply(_path, 576);

            Assert.That(File.ReadAllBytes(_path), Is.EqualTo(once));
        }

        // Verifies existing encoder colour information is preserved even when it differs from the fallback.
        [Test]
        public void Apply_PreservesExistingColorMetadata()
        {
            byte[] existingColor = Box("colr", Encoding.ASCII.GetBytes("existing-profile"));
            byte[] original = Movie(324, 576, existingColor);
            File.WriteAllBytes(_path, original);

            Mp4ColorMetadata.Apply(_path, 576);

            Assert.That(File.ReadAllBytes(_path), Is.EqualTo(original));
        }

        // Verifies 64-bit top-level media and movie sizes are parsed and retained correctly.
        [Test]
        public void Apply_AcceptsExtendedBoxSizes()
        {
            byte[] movie = Movie(324, 578);
            byte[] extendedMovie = ExtendedBox("moov", movie.Skip(8).ToArray());
            byte[] media = ExtendedBox("mdat", new byte[] { 12, 34 });
            File.WriteAllBytes(_path, Join(media, extendedMovie));

            Mp4ColorMetadata.Apply(_path, 578);

            byte[] result = File.ReadAllBytes(_path);
            Assert.That(result.Take(media.Length), Is.EqualTo(media));
            byte[] updated = result.Skip(media.Length + extendedMovie.Length).ToArray();
            Assert.That(updated.Take(4), Is.EqualTo(new byte[] { 0, 0, 0, 1 }));
            Assert.That(updated.Skip(8).Take(8), Is.EqualTo(BigEndian(updated.Length, 8)));
        }

        // Verifies unsupported or truncated files fail before changing any original byte.
        [TestCase("missing-movie")]
        [TestCase("truncated-header")]
        [TestCase("invalid-size")]
        [TestCase("open-ended-media")]
        [TestCase("missing-video")]
        public void Apply_RejectsInvalidFilesWithoutChangingThem(string scenario)
        {
            byte[] original = scenario switch
            {
                "missing-movie" => Box("mdat", new byte[] { 1 }),
                "truncated-header" => new byte[] { 0, 0, 0 },
                "invalid-size" => new byte[] { 0, 0, 0, 64, 109, 111, 111, 118 },
                "open-ended-media" => Join(Movie(324, 576), new byte[] { 0, 0, 0, 0, 109, 100, 97, 116 }),
                _ => Box("moov", Box("udta", new byte[] { 1 }))
            };
            File.WriteAllBytes(_path, original);

            Assert.Throws<InvalidDataException>(() => Mp4ColorMetadata.Apply(_path, 576));

            Assert.That(File.ReadAllBytes(_path), Is.EqualTo(original));
        }

        private static byte[] Movie(int width, int height, byte[] color = null)
        {
            byte[] header = new byte[78];
            Array.Copy(BigEndian(width, 2), 0, header, 24, 2);
            Array.Copy(BigEndian(height, 2), 0, header, 26, 2);
            byte[] configuration = Hex("0164001effe1001c6764001eac2cac15049e7f0110000003001000000303c0da0884538001000468ce3c30");
            byte[] entry = Box("avc1", header, Box("avcC", configuration), color ?? Array.Empty<byte>());
            byte[] description = Box("stsd", new byte[] { 0, 0, 0, 0, 0, 0, 0, 1 }, entry);
            byte[] offsets = Box("stco", new byte[] { 0, 0, 0, 0, 0, 0, 0, 1, 0, 0, 0, 8 });
            byte[] track = Box("trak", Box("mdia", Box("minf", Box("stbl", description, offsets))));
            // An audio track before the video track must not stop the path search.
            return Box("moov", Box("trak", Box("mdia", Box("hdlr", new byte[4]))), track,
                Box("udta", new byte[] { 93, 94 }));
        }

        private static byte[] Descend(byte[] root, params string[] path)
        {
            foreach (string type in path)
            {
                root = Children(root, Type(root) == "stsd" ? 16 : 8)
                    .Last(child => Type(child) == type);
                Assert.That(ReadSize(root), Is.EqualTo(root.Length));
            }

            return root;
        }

        private static System.Collections.Generic.IEnumerable<byte[]> Children(byte[] parent, int offset)
        {
            while (offset < parent.Length)
            {
                int length = ReadSize(parent, offset);
                yield return parent.Skip(offset).Take(length).ToArray();
                offset += length;
            }

            Assert.That(offset, Is.EqualTo(parent.Length));
        }

        private static int ReadSize(byte[] bytes, int offset = 0)
        {
            return (bytes[offset] << 24) | (bytes[offset + 1] << 16) | (bytes[offset + 2] << 8) | bytes[offset + 3];
        }

        private static string Type(byte[] box) => Encoding.ASCII.GetString(box, 4, 4);

        private static byte[] Box(string type, params byte[][] payload)
        {
            byte[] body = Join(payload);
            return Join(BigEndian(body.Length + 8, 4), Encoding.ASCII.GetBytes(type), body);
        }

        private static byte[] ExtendedBox(string type, byte[] payload)
        {
            return Join(new byte[] { 0, 0, 0, 1 }, Encoding.ASCII.GetBytes(type),
                BigEndian(payload.Length + 16, 8), payload);
        }

        private static byte[] BigEndian(int number, int length)
        {
            byte[] result = new byte[length];
            for (int i = length - 1; i >= 0; i--)
            {
                result[i] = (byte)number;
                number >>= 8;
            }

            return result;
        }

        private static byte[] Join(params byte[][] parts) => parts.SelectMany(part => part).ToArray();

        private static byte[] Hex(string text) => Enumerable.Range(0, text.Length / 2)
            .Select(index => Convert.ToByte(text.Substring(index * 2, 2), 16)).ToArray();
    }
}
