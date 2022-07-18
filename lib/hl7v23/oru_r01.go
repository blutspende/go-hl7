package hl7v23

// ORU_R01 - Unsolicited transmission of an observation message
// https://hl7-definition.caristix.com/v2/HL7v2.3/TriggerEvents/ORU_R01
type ORU_R01 struct {
	MSH MSH `hl7:"MSH"`

	PatientResult []struct {
		Patient struct {
			PatientIdentification PID   `hl7:"PID,optional"`
			PattientDemographics  PD1   `hl7:"PD1,optional"`
			NotesAndComments      []NTE `hl7:"NTE,optional"`
			Visit                 struct {
				PatientVisit          PV1 `hl7:"PV1,optional"`
				AdditionalInformation PV2 `hl7:"PV2,optional"`
			}
		}

		OrderObservation []struct {
			CommonOrder        ORC   `hl7:"ORC,optional"`
			ObservationRequest OBR   `hl7:"OBR,optional"`
			NotesAndComments   []NTE `hl7:"NTE,optional"`
			Observation        []struct {
				Observation                 OBX   `hl7:"OBX,optional"`
				ObservationNotesAndComments []NTE `hl7:"NTE,optional"`
			}
			ClinicalTrialIdentification []CTI `hl7:"CTI,optional"`
		}
	}
	ContinuationPointer DSC `hl7:"DSC"`
}
