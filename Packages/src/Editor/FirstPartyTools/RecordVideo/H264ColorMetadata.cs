using System;
using System.Collections.Generic;
using System.IO;
using System.Linq;

namespace io.github.hatayama.UnityCliLoop.FirstPartyTools
{
    /// <summary>
    /// Supplies missing SPS video signal information; slice data and timing syntax stay unchanged.
    /// </summary>
    internal static class H264ColorMetadata
    {
        internal static byte[] PatchConfiguration(byte[] configuration, int matrix, out byte[] colorBox)
        {
            using MemoryStream input = new MemoryStream(configuration, false);
            using MemoryStream output = new MemoryStream();
            if (configuration.Length < 7 || configuration[0] != 1)
            {
                throw new InvalidDataException("The MP4 AVC configuration is invalid.");
            }

            output.Write(configuration, 0, 6);
            input.Position = 6;
            int count = configuration[5] & 31;
            colorBox = null;
            for (int i = 0; i < count; i++)
            {
                int size = (ReadByte(input) << 8) | ReadByte(input);
                byte[] nal = new byte[size];
                if (input.Read(nal, 0, size) != size)
                {
                    throw new InvalidDataException("The MP4 SPS is truncated.");
                }

                byte[] patched = PatchSps(nal, matrix, out byte[] color);
                if (colorBox != null && !colorBox.SequenceEqual(color))
                {
                    throw new InvalidDataException("The MP4 contains conflicting SPS colour descriptions.");
                }

                colorBox = color;
                output.WriteByte(checked((byte)(patched.Length >> 8)));
                output.WriteByte((byte)patched.Length);
                output.Write(patched, 0, patched.Length);
            }

            if (colorBox == null)
            {
                throw new InvalidDataException("The MP4 AVC configuration has no SPS.");
            }

            input.CopyTo(output);
            return output.ToArray();
        }

        private static byte[] PatchSps(byte[] nal, int matrix, out byte[] colorBox)
        {
            if (nal.Length < 4 || (nal[0] & 31) != 7)
            {
                throw new InvalidDataException("The AVC configuration does not contain a valid SPS NAL.");
            }

            H264BitStream bits = new H264BitStream(nal);
            int profile = bits.Read(8);
            bits.Read(16);
            bits.ReadUnsignedExpGolomb();
            SkipProfile(bits, profile);
            bits.ReadUnsignedExpGolomb();
            SkipPictureOrder(bits);
            bits.ReadUnsignedExpGolomb();
            bits.Read(1);
            bits.ReadUnsignedExpGolomb();
            bits.ReadUnsignedExpGolomb();
            if (bits.Read(1) == 0)
            {
                bits.Read(1);
            }

            bits.Read(1);
            if (bits.Read(1) == 1)
            {
                for (int i = 0; i < 4; i++) bits.ReadUnsignedExpGolomb();
            }

            return PatchVui(bits, nal, matrix, out colorBox);
        }

        private static byte[] PatchVui(H264BitStream bits, byte[] nal, int matrix, out byte[] colorBox)
        {
            int vuiStart = bits.Position;
            bool hasVui = bits.Read(1) == 1;
            if (hasVui)
            {
                SkipAspectRatio(bits);
                if (bits.Read(1) == 1) bits.Read(1);
            }

            int signalStart = bits.Position;
            int format = 5;
            int fullRange = 0;
            int primaries = 2;
            int transfer = 2;
            bool hasDescription = false;
            if (hasVui && bits.Read(1) == 1)
            {
                format = bits.Read(3);
                fullRange = bits.Read(1);
                hasDescription = bits.Read(1) == 1;
                if (hasDescription)
                {
                    primaries = bits.Read(8);
                    transfer = bits.Read(8);
                    int existingMatrix = bits.Read(8);
                    if (existingMatrix != 2) matrix = existingMatrix;
                    else hasDescription = false;
                }
            }

            colorBox = new byte[] { 0, 0, 0, 19, 99, 111, 108, 114, 110, 99, 108, 120,
                0, (byte)primaries, 0, (byte)transfer, 0, (byte)matrix, (byte)(fullRange << 7) };
            if (hasDescription)
            {
                return nal;
            }

            List<bool> replacement = new List<bool>();
            if (!hasVui) H264BitStream.Append(replacement, 4, 3); // VUI present, no aspect or overscan.
            H264BitStream.Append(replacement, 1, 1);
            H264BitStream.Append(replacement, format, 3);
            H264BitStream.Append(replacement, fullRange, 1);
            H264BitStream.Append(replacement, 1, 1);
            H264BitStream.Append(replacement, primaries, 8);
            H264BitStream.Append(replacement, transfer, 8);
            H264BitStream.Append(replacement, matrix, 8);
            if (!hasVui)
            {
                H264BitStream.Append(replacement, 0, 6); // Chroma location, timing, HRD, pic structure, restriction.
                bits.Replace(vuiStart, 1, replacement);
            }
            else
            {
                bits.Replace(signalStart, bits.Position - signalStart, replacement);
            }

            return bits.ToNal(nal[0]);
        }

        private static void SkipProfile(H264BitStream bits, int profile)
        {
            int[] extendedProfiles = { 44, 83, 86, 100, 110, 118, 122, 128, 134, 135, 138, 139, 144, 244 };
            if (!extendedProfiles.Contains(profile)) return;
            int chromaFormat = bits.ReadUnsignedExpGolomb();
            if (chromaFormat == 3) bits.Read(1);
            bits.ReadUnsignedExpGolomb();
            bits.ReadUnsignedExpGolomb();
            bits.Read(1);
            if (bits.Read(1) == 0) return;
            for (int i = 0; i < (chromaFormat == 3 ? 12 : 8); i++)
            {
                if (bits.Read(1) == 1) SkipScalingList(bits, i < 6 ? 16 : 64);
            }
        }

        private static void SkipScalingList(H264BitStream bits, int size)
        {
            int lastScale = 8;
            int nextScale = 8;
            for (int i = 0; i < size; i++)
            {
                if (nextScale != 0) nextScale = (lastScale + bits.ReadSignedExpGolomb() + 256) % 256;
                if (nextScale != 0) lastScale = nextScale;
            }
        }

        private static void SkipPictureOrder(H264BitStream bits)
        {
            int type = bits.ReadUnsignedExpGolomb();
            if (type == 0) bits.ReadUnsignedExpGolomb();
            if (type != 1) return;
            bits.Read(1);
            bits.ReadSignedExpGolomb();
            bits.ReadSignedExpGolomb();
            int cycle = bits.ReadUnsignedExpGolomb();
            for (int i = 0; i < cycle; i++) bits.ReadSignedExpGolomb();
        }

        private static void SkipAspectRatio(H264BitStream bits)
        {
            if (bits.Read(1) == 1 && bits.Read(8) == 255) bits.Read(32);
        }

        private static int ReadByte(Stream stream)
        {
            int value = stream.ReadByte();
            if (value < 0) throw new InvalidDataException("The MP4 AVC configuration is truncated.");
            return value;
        }
    }
}
