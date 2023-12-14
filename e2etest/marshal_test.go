package e2e

import (
	"github.com/blutspende/go-hl7"
	"testing"
	"time"

	"github.com/blutspende/go-hl7/lib/hl7v23"
	"github.com/stretchr/testify/assert"
)

func TestMarshalMSH(t *testing.T) {
	// Arrange
	mshData := "MSH|^~\\&|HL7_Host|HL7_Office|CIT|LAB|20110926125155||ORM^O01|20110926125155|P|2.3|0||ER|ER||8859/1|\u000d"
	filedata := mshData

	var err error
	var message hl7v23.ORM_O01
	err = hl7.Unmarshal(
		[]byte(filedata),
		&message,
		hl7.EncodingUTF8,
		hl7.TimezoneEuropeBerlin)

	// Act
	var marshalledMessageBytes [][]byte
	marshalledMessageBytes, err = hl7.Marshal(
		message,
		hl7.StandardFieldSeparator,
		hl7.EncodingASCII,
		hl7.TimezoneEuropeBerlin,
		hl7.StandardNotation)

	// Assert
	assert.Nil(t, err)
	assert.Equal(t, mshData, string(marshalledMessageBytes[0]))
}

func TestMarshalFromStruct(t *testing.T) {
	mshData := "MSH|^~\\&|HL7_Host|HL7_Office|CIT|LAB|20230203080903||ORM^O01|CID586246|P|2.3|0||ER|SU||8859/1|\u000d"
	trigger := hl7v23.ORM_O01{}
	msh := hl7v23.MSH{}
	msh.ReceivingApplication = hl7v23.HD{NamespaceId: "CIT"}
	msh.ReceivingFacility = hl7v23.HD{NamespaceId: "LAB"}
	msh.SendingApplication = hl7v23.HD{NamespaceId: "HL7_Host"}
	msh.SendingFacility = hl7v23.HD{NamespaceId: "HL7_Office"}
	msh.DateTimeOfMessage, _ = time.Parse("20060102150405", "20230203080903")
	msh.MessageType = "ORM"
	msh.MessageTriggerEvent = "O01"
	msh.MessageControlID = "CID586246"
	msh.ProcessingID = "P"
	msh.VersionID = "2.3"
	msh.SequenceNumber = 0
	msh.AcceptAcknowledgementType = "ER"
	msh.ApplicationAcknowledgementType = "SU"
	msh.CharacterSet = []string{"8859/1"}
	trigger.MSH = msh

	trigger.Order = append(trigger.Order, hl7v23.Order{})
	{
		orc := &trigger.Order[0].CommonOrderSegment
		{
			orc.OrderControl = "NW"
			orc.FillerOrderNumber.EntityIdentifier = "23071012"
			orc.DateTimeOfTransaction = time.Now()
			orc.OrderingProvider.ID = "AKB"
		}
		orb := &trigger.Order[0].OrderDetail.OrderDetailSegment.ObservationRequestSegment
		{
			orb.ObservationRequest = "1"
			orb.FillerOrderNumber.EntityIdentifier = "FIL4345"
			orb.UniversalServiceIdentifier.Identifier = "DNA-A"
			orb.UniversalServiceIdentifier.Text = "Loki A am DNA Strang"
			orb.RequestedDateTime = time.Now()
		}
	}
	trigger.Order = append(trigger.Order, hl7v23.Order{})
	{
		orc := &trigger.Order[1].CommonOrderSegment
		{
			orc.OrderControl = "NW"
			orc.PlacerOrderNumber.EntityIdentifier = "23071012"
			orc.DateTimeOfTransaction = time.Now()
			orc.OrderingProvider.ID = "AKB"
		}
		orb := &trigger.Order[1].OrderDetail.OrderDetailSegment.ObservationRequestSegment
		{
			orb.ObservationRequest = "2"
			orb.PlacerOrderNumber.EntityIdentifier = "FIL4345"
			orb.UniversalServiceIdentifier.Identifier = "DNA-DRQB"
			orb.UniversalServiceIdentifier.Text = "DRQB Loki"
			orb.RequestedDateTime = time.Now()
		}
	}
	pid := &trigger.Patient.PatientIdentification
	{
		pid.InternalID = append(pid.InternalID, hl7v23.CX{})
		pid.InternalID[0].Id = "01020304"
		pid.Name.FamilyName = "Nachnamäh"
		pid.Name.GivenName = "Vörname"
		pid.Sex = "U"
	}
	pv1 := &trigger.Patient.PatientVisit.PatientVisit
	{
		pv1.VisitNumber.Id = "VID001"
		pv1.PatientClass = "S"
	}
	marshalledMessageBytes, err := hl7.Marshal(
		trigger,
		hl7.StandardFieldSeparator,
		hl7.EncodingASCII,
		hl7.TimezoneUTC,
		hl7.StandardNotation)
	assert.Nil(t, err)
	assert.Equal(t, mshData, string(marshalledMessageBytes[0]))

}

func TestMarshalPID(t *testing.T) {
	// Arrange
	mshData := "MSH|^~\\&|HL7_Host|HL7_Office|CIT|LAB|20110926125155||ORM^O01|20110926125155|P|2.3|0||ER|ER||8859/1|\u000d"
	pidData := "PID|1|a^b~^c|00100M56016||Smith^Harry||19500412|M\u000d"
	filedata := mshData + pidData

	var message hl7v23.ORM_O01
	err := hl7.Unmarshal(
		[]byte(filedata),
		&message,
		hl7.EncodingUTF8,
		hl7.TimezoneEuropeBerlin)

	// Act
	marshalledMessageBytes, err := hl7.Marshal(
		message,
		hl7.StandardFieldSeparator,
		hl7.EncodingASCII,
		hl7.TimezoneEuropeBerlin,
		hl7.StandardNotation)

	// Assert
	assert.Nil(t, err)
	assert.Equal(t, pidData, string(marshalledMessageBytes[1]))
}
