using System;
using UnityEditor;
using UnityEditor.Media;
using UnityEngine;

namespace io.github.hatayama.UnityCliLoop.FirstPartyTools
{
    /// <summary>
    /// Wraps UnityEditor.Media.MediaEncoder as IVideoFrameEncoder.
    /// </summary>
    internal sealed class MediaEncoderVideoFrameEncoder : IVideoFrameEncoder
    {
        private MediaEncoder _encoder;
        private readonly int _width;
        private readonly int _height;
        private readonly string _filePath;
        private readonly bool _useVp8;
        private bool _hasFrames;
        private bool _disposed;

        internal MediaEncoderVideoFrameEncoder(
            string filePath,
            int width,
            int height,
            int frameRate,
            bool useVp8,
            RecordVideoQuality quality)
        {
            Debug.Assert(!string.IsNullOrEmpty(filePath), "filePath must not be empty.");
            Debug.Assert(width > 0, "width must be positive.");
            Debug.Assert(height > 0, "height must be positive.");
            Debug.Assert(frameRate > 0, "frameRate must be positive.");

            _width = width;
            _height = height;
            _filePath = filePath;
            _useVp8 = useVp8;
            _encoder = new MediaEncoder(filePath, CreateAttributes(width, height, frameRate, useVp8, quality));
        }

        public int Width => _width;

        public int Height => _height;

        public bool AddFrame(Texture2D texture)
        {
            Debug.Assert(!_disposed, "AddFrame must not run after Dispose.");
            bool added = _encoder.AddFrame(texture);
            _hasFrames |= added;
            return added;
        }

        public void Dispose()
        {
            if (_disposed)
            {
                return;
            }

            _disposed = true;
            _encoder.Dispose();
            _encoder = null;
            if (Application.platform == RuntimePlatform.WindowsEditor && !_useVp8 && _hasFrames)
            {
                try
                {
                    Mp4ColorMetadata.Apply(_filePath, _height);
                }
                catch (Exception exception)
                {
                    // Metadata failure must not discard the finished recording or its frame evidence.
                    Debug.LogWarning($"Could not add MP4 colour metadata to '{_filePath}': {exception.Message}");
                }
            }
        }

        private static VideoTrackEncoderAttributes CreateAttributes(
            int width,
            int height,
            int frameRate,
            bool useVp8,
            RecordVideoQuality quality)
        {
            VideoTrackEncoderAttributes attributes = useVp8
                ? new VideoTrackEncoderAttributes(new VP8EncoderAttributes
                {
                    keyframeDistance = RecordVideoConstants.Vp8KeyframeDistance
                })
                : new VideoTrackEncoderAttributes(new H264EncoderAttributes
                {
                    gopSize = RecordVideoConstants.H264GopSize,
                    numConsecutiveBFrames = RecordVideoConstants.H264ConsecutiveBFrames,
                    profile = VideoEncodingProfile.H264High
                });

            attributes.frameRate = new MediaRational(frameRate);
            attributes.width = (uint)width;
            attributes.height = (uint)height;
            attributes.includeAlpha = false;
            attributes.bitRateMode = VideoQualityPolicy.ToBitrateMode(quality);
            return attributes;
        }
    }
}
