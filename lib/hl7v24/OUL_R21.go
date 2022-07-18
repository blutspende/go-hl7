package hl7v24

// HL7 v2.4 - OUL_R21 - Unsolicited laboratory observation
// https://hl7-definition.caristix.com/v2/HL7v2.4/TriggerEvents/OUL_R21
type OUL_R21 struct {
	MSH              MSH   `hl7:"MSH" json:"MSH,omitempty"`
	NotesAndComments []NTE `hl7:"NTE,optional" json:"NotesAndComments,omitempty"`
	Patient          struct {
		PatientIdentification PID   `hl7:"PID,optional" json:"PatientIdentification,omitempty"`
		PatientDemographics   PD1   `hl7:"PD1,optional" json:"PatientDemographics,omitempty"`
		NotesAndComments      []NTE `hl7:"NTE,optional" json:"NotesAndComments,omitempty"`
	}
	Visit struct {
		PatientVisit          PV1 `hl7:"PV1,optional" json:"PatientVisit,omitempty"`
		AdditionalInformation PV2 `hl7:"PV2,optional" json:"AdditionalInformation,omitempty"`
	}
	OrderObservation []struct {
		Container struct {
			SpecimenAndContainerDetail SAC   `hl7:"SAC,optional" json:"SpecimenAndContainerDetail,omitempty"`
			SubstanceIdentifier        SID   `hl7:"SID,optional" json:"SubstanceIdentifier,omitempty"`
			ObservationResult          []OBX `hl7:"OBX,optional" json:"ObservationResult,omitempty"`
		}
		CommonOrder        ORC   `hl7:"ORC,optional" json:"CommonOrder,omitempty"`
		ObservationRequest OBR   `hl7:"OBR,optional" json:"ObservationRequest,omitempty"`
		NotesAndComments   []NTE `hl7:"NTE,optional" json:"NotesAndComments,omitempty"`
		Observation        []struct {
			Observation         OBX   `hl7:"OBX,optional" json:"Observation,omitempty"`
			TestCodeDetail      TCD   `hl7:"TCD,optional" json:"TestCodeDetail,omitempty"`
			SubstanceIdentifier SID   `hl7:"SID,optional" json:"SubstanceIdentifier,omitempty"`
			NotesAndComments    []NTE `hl7:"NTE,optional" json:"NotesAndComments,omitempty"`
		}
		ClinicalTrialIdentification []CTI `hl7:"CTI,optional" json:"ClinicalTrialIdentification,omitempty"`
	}
	ContinuationPointer DSC `hl7:"DSC" json:"ContinuationPointer,omitempty"`
}
