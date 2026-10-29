package projectverify_test

import (
	"strings"
	"testing"

	"github.com/hatayama/unity-cli-loop/dispatcher/internal/projectverify"
)

// encodedGUID is a Tuanjie-encoded GUID taken from a real Tuanjie 1.9.1 project.
const encodedGUID = "BnIbtiz/Ui/6fZL//VTJXbIGZfPNbOHVCY7wmPCKA5sVsPPDJdpcv1M="

// encodedGUIDOtherCase is encodedGUID with one letter past the checked prefix in the other case,
// which is a different but still valid Tuanjie-encoded GUID.
const encodedGUIDOtherCase = "BnIbtiz/Ui/6fZL//VTJXbIGZfPNbOHVCY7wmPCKA5sVsPPDJdpcV1M="

// badPrefixEncodedGUID has the shape of a Tuanjie-encoded GUID but decodes to bytes Tuanjie never writes.
const badPrefixEncodedGUID = "AAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAAA="

// Verifies hex and Tuanjie-encoded GUIDs can sit side by side in one project without findings.
func TestRunAcceptsHexAndTuanjieEncodedGuidsTogether(t *testing.T) {
	root := writeProjectWithManifest(t, map[string]string{
		"Assets/Hex.txt":           "a",
		"Assets/Hex.txt.meta":      metaText(guidOf(1)),
		"Assets/Encoded.txt":       "b",
		"Assets/Encoded.txt.meta":  metaText(encodedGUID),
		"Assets/Encoded2.txt":      "c",
		"Assets/Encoded2.txt.meta": metaText("DylK5i+tBy3hwBpolH4EHssf3fnR6QNZrkcrOW/pENa5cIxYicJoEC4="),
	})

	assertFindings(t, runProject(t, root))
}

// Verifies two .meta files with the same Tuanjie-encoded GUID are one duplicate that names the GUID as written.
func TestRunReportsDuplicateTuanjieEncodedGuid(t *testing.T) {
	root := writeProjectWithManifest(t, map[string]string{
		"Assets/A.txt":      "a",
		"Assets/A.txt.meta": metaText(encodedGUID),
		"Assets/B.txt":      "b",
		"Assets/B.txt.meta": metaText(encodedGUID),
	})

	assertFindings(t, runProject(t, root), findingKey{
		Check:        projectverify.CheckGUIDDuplicate,
		Path:         "Assets/A.txt.meta",
		GUID:         encodedGUID,
		RelatedPaths: []string{"Assets/B.txt.meta"},
	})
}

// Verifies Tuanjie-encoded GUIDs that differ only in letter case are different GUIDs, since Base64 is
// case-sensitive.
func TestRunTreatsTuanjieEncodedGuidsAsCaseSensitive(t *testing.T) {
	if !strings.EqualFold(encodedGUID, encodedGUIDOtherCase) || encodedGUID == encodedGUIDOtherCase {
		t.Fatal("the two encoded GUIDs must differ only in letter case")
	}
	root := writeProjectWithManifest(t, map[string]string{
		"Assets/A.txt":      "a",
		"Assets/A.txt.meta": metaText(encodedGUID),
		"Assets/B.txt":      "b",
		"Assets/B.txt.meta": metaText(encodedGUIDOtherCase),
	})

	assertFindings(t, runProject(t, root))
}

// Verifies hex GUIDs that differ only in letter case are still one duplicate, keyed by the lower-case form.
func TestRunTreatsHexGuidsAsCaseInsensitive(t *testing.T) {
	root := writeProjectWithManifest(t, map[string]string{
		"Assets/A.txt":      "a",
		"Assets/A.txt.meta": metaText(strings.ToUpper(mixedCaseGUID)),
		"Assets/B.txt":      "b",
		"Assets/B.txt.meta": metaText(mixedCaseGUID),
	})

	assertFindings(t, runProject(t, root), findingKey{
		Check:        projectverify.CheckGUIDDuplicate,
		Path:         "Assets/A.txt.meta",
		GUID:         mixedCaseGUID,
		RelatedPaths: []string{"Assets/B.txt.meta"},
	})
}

// Verifies a value shaped like a Tuanjie-encoded GUID whose decoded prefix is wrong is GUID_INVALID
// with a message that points at a possible change in Tuanjie's encoding.
func TestRunReportsTuanjieEncodedGuidWithWrongPrefix(t *testing.T) {
	root := writeProjectWithManifest(t, map[string]string{
		"Assets/A.txt":      "a",
		"Assets/A.txt.meta": metaText(badPrefixEncodedGUID),
	})

	report := runProject(t, root)

	assertFindings(t, report, guidInvalid("Assets/A.txt.meta"))
	assertMessagesAt(t, report, projectverify.CheckGUIDInvalid, "Assets/A.txt.meta",
		"Assets/A.txt.meta has a guid that looks like a Tuanjie-encoded GUID but does not decode the way"+
			" Tuanjie's encoding does, so the value is probably damaged. If every Tuanjie-encoded GUID in the"+
			" project is reported like this, this Tuanjie version probably encodes GUIDs differently and"+
			" verify-project needs updating.")
}

// Verifies a Tuanjie-encoded GUID in a .meta file with CRLF line endings is read without the
// carriage return.
func TestRunReadsTuanjieEncodedGuidWithCRLF(t *testing.T) {
	root := writeProjectWithManifest(t, map[string]string{
		"Assets/A.txt":      "a",
		"Assets/A.txt.meta": "fileFormatVersion: 2\r\nguid: " + encodedGUID + "\r\n",
		"Assets/B.txt":      "b",
		"Assets/B.txt.meta": "fileFormatVersion: 2\r\nguid: " + encodedGUID + "\r\n",
	})

	assertFindings(t, runProject(t, root), findingKey{
		Check:        projectverify.CheckGUIDDuplicate,
		Path:         "Assets/A.txt.meta",
		GUID:         encodedGUID,
		RelatedPaths: []string{"Assets/B.txt.meta"},
	})
}
