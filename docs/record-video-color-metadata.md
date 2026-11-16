# Windows recording colour metadata

`record-video` still captures RGBA frames and encodes them with Unity's `MediaEncoder`.
On Windows, completed H.264 MP4 recordings receive missing colour metadata before the
recording is returned. No external program is required for this step. VP8 and other
platforms keep their existing behaviour.

The measured Windows encoder conversion uses limited range and chooses its matrix from
the encoded video height, independently of width:

- Height greater than 576: BT.709 (`matrix_coefficients=1`).
- Height at most 576: BT.601 (`matrix_coefficients=6`, SMPTE 170M).

This agrees with Microsoft's documented default in
[DXVA2_VideoTransferMatrix](https://learn.microsoft.com/en-us/windows/win32/api/dxva2api/ne-dxva2api-dxva2_videotransfermatrix).
It describes the tested Unity Windows encoder path, not a guarantee about every hardware
encoder. Recheck the conversion if Unity changes its encoder backend. Existing explicit
SPS colour information or an existing MP4 `colr` box is preserved. Primaries and transfer
remain unspecified when the encoder did not supply them; a YUV matrix does not establish
either of those properties.

The patch writes both the SPS VUI in `avcC` and a matching MP4 `colr`/`nclx` box.
Container metadata alone fixes FFmpeg extraction, but the Windows Media Foundation
decoder needs the SPS information. Its output media type exposes that information after
the first decoded sample, rather than before decoding starts.

Only the small `moov` metadata tree is rewritten. The new tree is appended and the old
tree becomes a `free` box, leaving compressed samples, timestamps, and absolute chunk
offsets unchanged. The extra space is one original metadata tree plus the colour fields;
the operation does not copy or recompress the media payload. A patch failure logs a warning
and keeps the finished recording and its frame evidence.

`Mp4ColorMetadataTests` and `H264ColorMetadataTests` cover the height boundary, metadata
preservation, SPS syntax, box sizes, unchanged media data, malformed inputs, and repeated
application. Live verification should additionally encode the same RGB colour chart at
750x1334, 336x600, 324x578, 324x576, 300x532, 186x332, and 1280x480, then compare default
decoded RGB against the source and explicitly selected matrices. Check compressed packet
hashes and timestamps before and after patching, and check the native Windows decoder's
matrix/range after decoding its first sample. FFmpeg is useful for these development checks
but remains optional at recording time.

## Validation on 2026-11-16

Tuanjie 1.9.1 (`2022.3.62t9`) on Windows, H.264 High, 30 frames per RGB colour chart.
Mean absolute RGB errors below use interior pixels where choosing the other matrix changes
at least one channel by six levels. The default extraction matched an explicitly selected
correct matrix byte for byte.

| Encoded size | Matrix | Default extraction error | Wrong matrix error |
| --- | --- | --- | --- |
| 750x1334 | BT.709 | 1.4866 | 11.5971 |
| 336x600 | BT.709 | 1.2834 | 11.5172 |
| 324x578 | BT.709 | 1.2786 | 11.5843 |
| 324x576 | BT.601 | 1.7167 | 11.1161 |
| 300x532 | BT.601 | 1.5997 | 10.4885 |
| 186x332 | BT.601 | 1.4508 | 10.8562 |
| 1280x480 | BT.601 | 1.2882 | 10.8030 |

All seven files retained identical compressed packet hashes, PTS, DTS, durations, and frame
counts. They decoded fully without errors. The native Media Foundation decoder reported
the expected matrix and limited range for every size after its first decoded sample.
An actual `record-video --seconds 1 --window-name Scene` recording returned 35 encoded frames,
its timeline PNG and analysis JSON; `--video ... --frame-at 0.5` extracted a PNG successfully.
The scoped EditMode run passed 74 tests across the two new metadata fixtures and the existing
recording session, session host, and analysis fixtures. Compilation had no errors or warnings;
the production complexity and file-length checks passed.
