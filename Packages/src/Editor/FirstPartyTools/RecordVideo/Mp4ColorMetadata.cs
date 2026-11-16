using System;
using System.Collections.Generic;
using System.IO;
using System.Text;

namespace io.github.hatayama.UnityCliLoop.FirstPartyTools
{
    /// <summary>
    /// Adds missing MP4 colour information without changing encoded frames or chunk offsets.
    /// </summary>
    internal static class Mp4ColorMetadata
    {
        private static readonly string[] VideoPath = { "moov", "trak", "mdia", "minf", "stbl", "stsd", "avc1" };

        internal static void Apply(string filePath, int height)
        {
            using FileStream file = new FileStream(filePath, FileMode.Open, FileAccess.ReadWrite, FileShare.None);
            Box movie = FindMovie(file);
            byte[] original = new byte[checked((int)movie.Size)];
            file.Position = movie.Start;
            ReadExactly(file, original);
            byte[] updated = AddColorBox(original, height);
            if (ReferenceEquals(original, updated))
            {
                return;
            }

            // Keep mdat and its absolute chunk offsets in place, including fast-start files.
            // Publish the new moov only after writing it; the old moov becomes padding.
            long originalLength = file.Length;
            try
            {
                file.Position = originalLength;
                file.Write(updated, 0, updated.Length);
                file.Flush();
            }
            catch
            {
                file.SetLength(originalLength);
                throw;
            }

            file.Position = movie.Start + 4;
            file.Write(Encoding.ASCII.GetBytes("free"), 0, 4);
        }

        private static Box FindMovie(Stream stream)
        {
            Box? movie = null;
            while (stream.Position < stream.Length)
            {
                Box box = ReadBox(stream, stream.Length);
                if (box.Type == "moov")
                {
                    movie = box;
                }

                if (box.ExtendsToEnd)
                {
                    throw new InvalidDataException("Cannot append colour metadata after an open-ended MP4 box.");
                }

                stream.Position = box.End;
            }

            return movie ?? throw new InvalidDataException("The MP4 has no moov box.");
        }

        private static byte[] AddColorBox(byte[] movie, int height)
        {
            using MemoryStream input = new MemoryStream(movie, false);
            List<Box> path = new List<Box>();
            if (!FindVideoPath(input, movie.Length, 0, path))
            {
                throw new InvalidDataException("The MP4 has no H.264 video sample entry.");
            }

            Box video = path[path.Count - 1];
            Box? configuration = null;
            input.Position = video.Start + video.HeaderSize + 78;
            while (input.Position < video.End)
            {
                Box child = ReadBox(input, video.End);
                if (child.Type == "colr")
                {
                    return movie;
                }

                if (child.Type == "avcC") configuration = child;

                input.Position = child.End;
            }

            Box avc = configuration ?? throw new InvalidDataException("The MP4 has no AVC configuration.");
            byte[] payload = new byte[checked((int)avc.Size - avc.HeaderSize)];
            input.Position = avc.Start + avc.HeaderSize;
            ReadExactly(input, payload);
            byte[] patched = H264ColorMetadata.PatchConfiguration(payload, height > 576 ? 1 : 6, out byte[] color);
            using MemoryStream output = new MemoryStream();
            output.Write(movie, 0, checked((int)avc.Start + avc.HeaderSize));
            output.Write(patched, 0, patched.Length);
            output.Write(movie, (int)avc.End, checked((int)(video.End - avc.End)));
            output.Write(color, 0, color.Length);
            output.Write(movie, (int)video.End, movie.Length - (int)video.End);
            int growth = patched.Length - payload.Length + color.Length;
            WriteBoxSize(output, avc, avc.HeaderSize + patched.Length);
            foreach (Box box in path)
            {
                WriteBoxSize(output, box, box.Size + growth);
            }

            return output.ToArray();
        }

        private static void WriteBoxSize(Stream output, Box box, long size)
        {
            output.Position = box.Start + (box.HeaderSize == 16 ? 8 : 0);
            WriteBigEndian(output, checked((ulong)size), box.HeaderSize == 16 ? 8 : 4);
        }

        private static bool FindVideoPath(Stream stream, long end, int depth, List<Box> path)
        {
            while (stream.Position < end)
            {
                Box box = ReadBox(stream, end);
                if (box.Type == VideoPath[depth])
                {
                    path.Add(box);
                    if (depth == VideoPath.Length - 1)
                    {
                        if (box.Size < box.HeaderSize + 78)
                        {
                            throw new InvalidDataException("The MP4 video sample entry is truncated.");
                        }

                        return true;
                    }

                    stream.Position = box.Start + box.HeaderSize + (box.Type == "stsd" ? 8 : 0);
                    if (FindVideoPath(stream, box.End, depth + 1, path))
                    {
                        return true;
                    }

                    path.RemoveAt(path.Count - 1);
                }

                stream.Position = box.End;
            }

            return false;
        }

        private static Box ReadBox(Stream stream, long end)
        {
            long start = stream.Position;
            if (end - start < 8)
            {
                throw new InvalidDataException("The MP4 box header is truncated.");
            }

            ulong size = ReadBigEndian(stream, 4);
            byte[] type = new byte[4];
            ReadExactly(stream, type);
            bool extendsToEnd = size == 0;
            int headerSize = size == 1 ? 16 : 8;
            if (headerSize == 16)
            {
                size = ReadBigEndian(stream, 8);
            }

            if (extendsToEnd)
            {
                size = (ulong)(end - start);
            }

            if (size < (ulong)headerSize || size > (ulong)(end - start))
            {
                throw new InvalidDataException("The MP4 box size exceeds its container.");
            }

            return new Box(start, (long)size, headerSize, Encoding.ASCII.GetString(type), extendsToEnd);
        }

        private static ulong ReadBigEndian(Stream stream, int count)
        {
            ulong value = 0;
            for (int i = 0; i < count; i++)
            {
                int next = stream.ReadByte();
                if (next < 0)
                {
                    throw new EndOfStreamException();
                }

                value = (value << 8) | (byte)next;
            }

            return value;
        }

        private static void WriteBigEndian(Stream stream, ulong value, int count)
        {
            for (int i = count - 1; i >= 0; i--)
            {
                stream.WriteByte((byte)(value >> (i * 8)));
            }
        }

        private static void ReadExactly(Stream stream, byte[] buffer)
        {
            int offset = 0;
            while (offset < buffer.Length)
            {
                int count = stream.Read(buffer, offset, buffer.Length - offset);
                if (count == 0)
                {
                    throw new EndOfStreamException();
                }

                offset += count;
            }
        }

        private readonly struct Box
        {
            internal readonly long Start;
            internal readonly long Size;
            internal readonly int HeaderSize;
            internal readonly string Type;
            internal readonly bool ExtendsToEnd;
            internal long End => Start + Size;

            internal Box(long start, long size, int headerSize, string type, bool extendsToEnd)
            {
                Start = start;
                Size = size;
                HeaderSize = headerSize;
                Type = type;
                ExtendsToEnd = extendsToEnd;
            }
        }
    }
}
