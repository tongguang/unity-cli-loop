using System.Collections.Generic;
using System.IO;

namespace io.github.hatayama.UnityCliLoop.FirstPartyTools
{
    /// <summary>
    /// Reads SPS syntax and replaces bits while preserving RBSP trailing bits and NAL escaping.
    /// </summary>
    internal sealed class H264BitStream
    {
        private readonly List<bool> _bits = new List<bool>();
        internal int Position { get; private set; }

        internal H264BitStream(byte[] nal)
        {
            int zeros = 0;
            for (int i = 1; i < nal.Length; i++)
            {
                byte value = nal[i];
                if (zeros == 2 && value == 3)
                {
                    zeros = 0;
                    continue;
                }

                Append(_bits, value, 8);
                zeros = value == 0 ? zeros + 1 : 0;
            }

            // Keep rbsp_stop_one_bit, but regenerate alignment after inserting VUI bits.
            while (_bits.Count > 0 && !_bits[_bits.Count - 1])
            {
                _bits.RemoveAt(_bits.Count - 1);
            }
        }

        internal int Read(int count)
        {
            if (count < 0 || count > 32 || Position + count > _bits.Count)
            {
                throw new InvalidDataException("The H.264 SPS is truncated.");
            }

            int value = 0;
            for (int i = 0; i < count; i++)
            {
                value = (value << 1) | (_bits[Position++] ? 1 : 0);
            }

            return value;
        }

        internal int ReadUnsignedExpGolomb()
        {
            int zeros = 0;
            while (Read(1) == 0)
            {
                if (++zeros > 30)
                {
                    throw new InvalidDataException("The H.264 SPS Exp-Golomb value is too large.");
                }
            }

            return ((1 << zeros) - 1) + Read(zeros);
        }

        internal int ReadSignedExpGolomb()
        {
            int value = ReadUnsignedExpGolomb();
            return (value & 1) == 0 ? -(value / 2) : (value + 1) / 2;
        }

        internal void Replace(int start, int count, List<bool> replacement)
        {
            _bits.RemoveRange(start, count);
            _bits.InsertRange(start, replacement);
        }

        internal byte[] ToNal(byte header)
        {
            List<byte> result = new List<byte> { header };
            int zeros = 0;
            for (int start = 0; start < _bits.Count; start += 8)
            {
                int value = 0;
                for (int bit = 0; bit < 8; bit++)
                {
                    value = (value << 1) | (start + bit < _bits.Count && _bits[start + bit] ? 1 : 0);
                }

                if (zeros == 2 && value <= 3)
                {
                    result.Add(3);
                    zeros = 0;
                }

                result.Add((byte)value);
                zeros = value == 0 ? zeros + 1 : 0;
            }

            return result.ToArray();
        }

        internal static void Append(List<bool> bits, int value, int count)
        {
            for (int i = count - 1; i >= 0; i--)
            {
                bits.Add(((value >> i) & 1) != 0);
            }
        }
    }
}
