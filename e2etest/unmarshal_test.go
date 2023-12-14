package e2e

import (
	"encoding/json"
	"fmt"
	"github.com/blutspende/go-hl7"
	"os"
	"testing"

	"github.com/blutspende/go-hl7/lib/hl7v23"
	"github.com/stretchr/testify/assert"
)

func TestMessageIdentification(t *testing.T) {
	data := "MSH|^~\\&|HL7_Host|HL7_Office|CIT|LAB|20110926125155||ORM^O01|20110926125155|P|2.3"

	messageType, protocolVersion, err := hl7.IdentifyMessage(
		[]byte(data),
		hl7.EncodingUTF8,
	)

	assert.Nil(t, err)
	assert.Equal(t, "ORM^O01", messageType)
	assert.Equal(t, "2.3", protocolVersion)
}

// Test_Parse_MSH_Segment, this test has only one line
func Test_Parse_MSH_Segment(t *testing.T) {
	fileData := fmt.Sprintf("MSH|^~\\&|HL7_Host^b^c|HL7_Office^^Xyz|CIT^^|LAB|20110926125155||ORM^O01|20110926125155|P|2.3|||ER|ER||8859/1~second_element|<\u000d")

	var message hl7v23.ORM_O01
	err := hl7.Unmarshal(
		[]byte(fileData),
		&message,
		hl7.EncodingUTF8,
		hl7.TimezoneEuropeBerlin)

	assert.Nil(t, err)
	assert.NotNil(t, message.MSH)
	assert.Equal(t, "|", message.MSH.FieldSeparator)
	assert.Equal(t, "^~\\&", message.MSH.EncodingCharacters)
	assert.Equal(t, "HL7_Host", message.MSH.SendingApplication.NamespaceId)
	assert.Equal(t, "b", message.MSH.SendingApplication.UniversalId)
	assert.Equal(t, "c", message.MSH.SendingApplication.UniversalIdType)
	assert.Equal(t, "HL7_Office", message.MSH.SendingFacility.NamespaceId)
	assert.Equal(t, "CIT", message.MSH.ReceivingApplication.NamespaceId)
	assert.Equal(t, "LAB", message.MSH.ReceivingFacility.NamespaceId)
	assert.Equal(t, "2011-09-26 10:51:55 +0000 UTC", message.MSH.DateTimeOfMessage.String())
	assert.Equal(t, "", message.MSH.Security)
	assert.Equal(t, "ORM", message.MSH.MessageType)
	assert.Equal(t, "O01", message.MSH.MessageTriggerEvent)
	assert.Equal(t, "20110926125155", message.MSH.MessageControlID)
	assert.Equal(t, "P", message.MSH.ProcessingID)
	assert.Equal(t, "2.3", message.MSH.VersionID)
	assert.Equal(t, 0, message.MSH.SequenceNumber)
	assert.Equal(t, "", message.MSH.ContinuationPointer)
	assert.Equal(t, "ER", message.MSH.AcceptAcknowledgementType)
	assert.Equal(t, "ER", message.MSH.ApplicationAcknowledgementType)
	assert.Equal(t, "", message.MSH.CountryCode)
	assert.Equal(t, "8859/1", message.MSH.CharacterSet[0])
	assert.Equal(t, "second_element", message.MSH.CharacterSet[1])
	assert.Equal(t, "<", message.MSH.PrincipalLanguageOfMessage.Identifier)
}

// Run a Testorder provided by Roche cITM but its some standard with each record once
func Test_Order_ORM_generic1(t *testing.T) {
	var filedata string
	filedata = filedata + "MSH|^~\\&|HL7_Host|HL7_Office|CIT|LAB|20110926125155||ORM^O01|20110926125155|P|2.3|||ER|ER||8859/1|\u000d"
	filedata = filedata + "PID|1|a^b~^c|00100M56016||Smith^Harry||19500412|M\u000d"
	filedata = filedata + "ORC|NW|000218T018||||Not used|^^^^^R||20110926120055\u000d"
	filedata = filedata + "OBR|1|000218T018||101~102||20110926120000|||||A||||\u000d"

	var message hl7v23.ORM_O01
	err := hl7.Unmarshal(
		[]byte(filedata),
		&message,
		hl7.EncodingUTF8,
		hl7.TimezoneEuropeBerlin)

	assert.Nil(t, err)

	assert.NotNil(t, message.Patient)
	assert.NotNil(t, message.Patient.PatientIdentification)
	assert.Equal(t, 1, message.Patient.PatientIdentification.ID)
	assert.Equal(t, "a", message.Patient.PatientIdentification.ExternalID[0].Id)
	assert.Equal(t, "b", message.Patient.PatientIdentification.ExternalID[0].CheckDigit)
	assert.Equal(t, "", message.Patient.PatientIdentification.ExternalID[0].AssigningAuthority)
	assert.Equal(t, "", message.Patient.PatientIdentification.ExternalID[1].Id)
	assert.Equal(t, "c", message.Patient.PatientIdentification.ExternalID[1].CheckDigit)
	assert.Equal(t, "", message.Patient.PatientIdentification.ExternalID[1].AssigningAuthority)
	assert.Equal(t, "a", message.Patient.PatientIdentification.ExternalID[0].Id)
	assert.Equal(t, "b", message.Patient.PatientIdentification.ExternalID[0].CheckDigit)
	assert.Equal(t, "", message.Patient.PatientIdentification.ExternalID[1].Id)
	assert.Equal(t, "c", message.Patient.PatientIdentification.ExternalID[1].CheckDigit)
	assert.Equal(t, "00100M56016", message.Patient.PatientIdentification.InternalID[0].Id)
	assert.Equal(t, "", message.Patient.PatientIdentification.AlternateID[0].Id)
	assert.Equal(t, "Smith", message.Patient.PatientIdentification.Name.FamilyName)
	assert.Equal(t, "Harry", message.Patient.PatientIdentification.Name.GivenName)
	assert.Equal(t, "", message.Patient.PatientIdentification.Name.MiddleInitialOrName)
	assert.Equal(t, "", message.Patient.PatientIdentification.MothersMaidenName.FamilyName)
	assert.Equal(t, "1950-04-12 00:00:00 +0100 CET", message.Patient.PatientIdentification.DOB.String())
	assert.Equal(t, "M", message.Patient.PatientIdentification.Sex)

	assert.Equal(t, 1, len(message.Order))

	assert.Equal(t, "NW", message.Order[0].CommonOrderSegment.OrderControl)
	assert.Equal(t, "000218T018", message.Order[0].CommonOrderSegment.PlacerOrderNumber.EntityIdentifier)
	assert.Equal(t, "", message.Order[0].CommonOrderSegment.PlacerOrderNumber.NamespaceId)
	assert.Equal(t, "", message.Order[0].CommonOrderSegment.PlacerOrderNumber.UniversalId)
	assert.Equal(t, "", message.Order[0].CommonOrderSegment.PlacerOrderNumber.UniversalIdType)
	assert.Equal(t, "", message.Order[0].CommonOrderSegment.FillerOrderNumber.EntityIdentifier)
	assert.Equal(t, "", message.Order[0].CommonOrderSegment.PlacerGroupNumber.EntityIdentifier)
	assert.Equal(t, "", message.Order[0].CommonOrderSegment.OrderStatus)
	assert.Equal(t, "Not used", message.Order[0].CommonOrderSegment.ResponseFlag)
	assert.Equal(t, "R", message.Order[0].CommonOrderSegment.QuantityTiming.Priority)
	assert.Equal(t, "", message.Order[0].CommonOrderSegment.QuantityTiming.Duration)
	assert.Equal(t, "R", message.Order[0].CommonOrderSegment.QuantityTiming.Priority)
	assert.Equal(t, "", message.Order[0].CommonOrderSegment.QuantityTiming.Condition)
	assert.Equal(t, "", message.Order[0].CommonOrderSegment.ParentOrder.ParentsPlacerOrderNumber)
	assert.Equal(t, "", message.Order[0].CommonOrderSegment.ParentOrder.ParentsFillerOrderNumber)
	assert.Equal(t, "2011-09-26 10:00:55 +0000 UTC", message.Order[0].CommonOrderSegment.DateTimeOfTransaction.String())

	assert.Nil(t, err)
	assert.NotNil(t, message.Order[0].OrderDetail.OrderDetailSegment.ObservationRequestSegment)
	assert.Equal(t, "1", message.Order[0].OrderDetail.OrderDetailSegment.ObservationRequestSegment.ObservationRequest)
	assert.Equal(t, "000218T018", message.Order[0].OrderDetail.OrderDetailSegment.ObservationRequestSegment.PlacerOrderNumber.EntityIdentifier)
	assert.Equal(t, "", message.Order[0].OrderDetail.OrderDetailSegment.ObservationRequestSegment.PlacerOrderNumber.NamespaceId)
	assert.Equal(t, "", message.Order[0].OrderDetail.OrderDetailSegment.ObservationRequestSegment.PlacerOrderNumber.UniversalId)
	assert.Equal(t, "", message.Order[0].OrderDetail.OrderDetailSegment.ObservationRequestSegment.PlacerOrderNumber.UniversalIdType)
	assert.Equal(t, "", message.Order[0].OrderDetail.OrderDetailSegment.ObservationRequestSegment.FillerOrderNumber.EntityIdentifier)
	assert.Equal(t, "101", message.Order[0].OrderDetail.OrderDetailSegment.ObservationRequestSegment.UniversalServiceIdentifier.Identifier)
	assert.Equal(t, "", message.Order[0].OrderDetail.OrderDetailSegment.ObservationRequestSegment.UniversalServiceIdentifier.Text)
	assert.Equal(t, "", message.Order[0].OrderDetail.OrderDetailSegment.ObservationRequestSegment.Priority)
	assert.Equal(t, "2011-09-26 10:00:00 +0000 UTC", message.Order[0].OrderDetail.OrderDetailSegment.ObservationRequestSegment.RequestedDateTime.String())
	assert.Equal(t, "0001-01-01 00:00:00 +0000 UTC", message.Order[0].OrderDetail.OrderDetailSegment.ObservationRequestSegment.ObservationDateTime.String())
	assert.Equal(t, "A", message.Order[0].OrderDetail.OrderDetailSegment.ObservationRequestSegment.SpecimenActionCode)

}

// See it yourself: uncomment this to locally check the HL7 struct format (exported to JSON)
// Open the JSON with your favorite editor, fold the segments for an easy read.
func TestMSH(t *testing.T) {
	sample := `MSH|^~\&|SWISSLAB|FFM||FFM|20230203080903||ORM^O01|o3057937.000001|P|2.3|` + "\r"
	sample += `PID|||01077843||CL5AVA0N7K|||U|` + "\r"
	sample += `PV1||S|||||||||||||||||S01077843|` + "\r"
	sample += `ORC|NW||23071012||||||20230203080800|||BSD|` + "\r"
	sample += `OBR|||23071012|HINAET^HIV-1/2 PCR mit erhöhter Sensitivität|||20230203080900|||||||||BSD|||||||||P|` + "\r"
	sample += `ORC|NW||23071013||||||20230203080800|||BSD|` + "\r"
	sample += `OBR|||23071013|HCNAET^HCV PCR mit erhöhter Sensitivität|||20230203080900|||||||||BSD|||||||||P|` + "\r"
	sample += `ORC|NW||23071014||||||20230203080800|||BSD|` + "\r"
	sample += `OBR|||23071014|HBNAET^HBV PCR mit erhöhter Sensitivität|||20230203080900|||||||||BSD|||||||||P|` + "\r"
	dest := hl7v23.ORM_O01{}
	err := hl7.Unmarshal([]byte(sample), &dest, hl7.EncodingUTF8, hl7.TimezoneEuropeBerlin)
	out, _ := json.MarshalIndent(dest, "", "\t")
	assert.Nil(t, err)
	file, err := os.Create("testMSH.json")
	if err != nil {
		t.Error(err)
	} else {
		file.Write(out)
		file.Close()
	}
}
