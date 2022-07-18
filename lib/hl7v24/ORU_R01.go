package hl7v24

// HL7 v2.4 - ORU_R01 - Unsolicited transmission of an observation message
// https://hl7-definition.caristix.com/v2/HL7v2.4/TriggerEvents/ORU_R01
type ORU_R01 struct {
	MSH MSH `hl7:"MSH" json:"MSH,omitempty"`

	PatientResult []struct {
		Patient struct {
			PatientIdentification PID   `hl7:"PID,optional" json:"PatientIdentification,omitempty"`
			PatientDemographics   PD1   `hl7:"PD1,optional" json:"PatientDemographics,omitempty"`
			NextOfKin             NK1   `hl7:"PD1,optional" json:"NextOfKin,omitempty"`
			NotesAndComments      []NTE `hl7:"NTE,optional" json:"NotesAndComments,omitempty"`
			Visit                 struct {
				PatientVisit          PV1 `hl7:"PV1,optional" json:"PatientVisit,omitempty"`
				AdditionalInformation PV2 `hl7:"PV2,optional" json:"AdditionalInformation,omitempty"`
			}
		}

		OrderObservation []struct {
			CommonOrder        ORC   `hl7:"ORC,optional" json:"CommonOrder,omitempty"`
			ObservationRequest OBR   `hl7:"OBR,optional" json:"ObservationRequest,omitempty"`
			NotesAndComments   []NTE `hl7:"NTE,optional" json:"NotesAndComments,omitempty"`
			ContactData        []CTD `hl7:"CTD,optional" json:"ContactData,omitempty"`

			Observation []struct {
				Observation                 OBX   `hl7:"OBX,optional" json:"Observation,omitempty"`
				ObservationNotesAndComments []NTE `hl7:"NTE,optional" json:"ObservationNotesAndComments,omitempty"`
			}
			FinancialTransaction        []FT1 `hl7:"FT1,optional" json:"FinancialTransaction,omitempty"`
			ClinicalTrialIdentification []CTI `hl7:"CTI,optional" json:"ClinicalTrialIdentification,omitempty"`
		}
	}
	ContinuationPointer DSC `hl7:"DSC" json:"ContinuationPointer,omitempty"`
}
