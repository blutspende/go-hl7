package hl7v24

// HL7 v2.4 - OML_O21 - Laboratory order
// https://hl7-definition.caristix.com/v2/HL7v2.4/TriggerEvents/OML_O21
type OML_O21 struct {
	MSH              MSH   `hl7:"MSH" json:"MSH"`
	NotesAndComments []NTE `hl7:"NTE,optional" json:"NotesAndComments"`
	Patient          struct {
		PatientIdentification         PID   `hl7:"PID" json:"PatientIdentification"`
		PatientAdditionalDemographics PD1   `hl7:"PD1,optional" json:"PatientAdditionalDemographics"`
		NotesAndComments              []NTE `hl7:"NTE,optional" json:"NotesAndComments"`
		Visit                         struct {
			PatientVisit          PV1 `hl7:"PV1" json:"PatientVisit"`
			AdditionalInformation PV2 `hl7:"PV2,optional" json:"AdditionalInformation"`
		}
		Insurance []struct {
			Insurance                          IN1 `hl7:"IN1" json:"Insurance"`
			AdditionalInformation              IN2 `hl7:"IN2,optional" json:"AdditionalInformation"`
			AdditionalInformationCertification IN3 `hl7:"IN3,optional" json:"AdditionalInformationCertification"`
		}
		Guarantor                 GT1   `hl7:"GT1,optional" json:"Guarantor"`
		PatientAllergyInformation []AL1 `hl7:"AL1,optional" json:"PatientAllergyInformation"`
	}
	OrderGeneral []struct {
		Container struct {
			SpecimenAndContainerDetail SAC   `hl7:"SAC" json:"SpecimenAndContainerDetail"`
			ObservationResult          []OBX `hl7:"OBX,optional" json:"ObservationResult"`
		}
		Order []struct {
			CommonOrder        ORC `hl7:"ORC" json:"CommonOrder"`
			ObservationRequest struct {
				ObservationRequest OBR `hl7:"OBR" json:"ObservationRequest"`
				Container          []struct {
					TestCodeDetail   TCD   `hl7:"TCD,optional" json:"TestCodeDetail"`
					NotesAndComments []NTE `hl7:"NTE,optional" json:"NotesAndComments"`
					Diagnosis        []DG1 `hl7:"DG1,optional" json:"Diagnosis"`
				}
				Observation []struct {
					ObservationResult OBX   `hl7:"OBX" json:"ObservationResult"`
					TestCodeDetail    TCD   `hl7:"TCD,optional" json:"TestCodeDetail"`
					NotesAndComments  []NTE `hl7:"NTE,optional" json:"NotesAndComments"`
				}
				PriorResult []struct {
					PatientPrior struct {
						PatientIdentification         PID `hl7:"PID" json:"PatientIdentification"`
						PatientAdditionalDemographics PD1 `hl7:"PD1,optional" json:"PatientAdditionalDemographics"`
					}
					PatientVisitPrior struct {
						PatientVisit          PV1 `hl7:"PV1" json:"PatientVisit"`
						AdditionalInformation PV2 `hl7:"PV2,optional" json:"AdditionalInformation"`
					}
					PatientAllergyInformation []AL1 `hl7:"AL1,optional" json:"PatientAllergyInformation"`
					OrderPrior                []struct {
						CommonOrder        ORC   `hl7:"ORC,optional" json:"CommonOrder"`
						ObservationRequest OBR   `hl7:"OBR" json:"ObservationRequest"`
						NotesAndComments   []NTE `hl7:"NTE,optional" json:"NotesAndComments"`
						ObservationPrior   []struct {
							ObservationResult OBX   `hl7:"OBX" json:"ObservationResult"`
							NotesAndComments  []NTE `hl7:"NTE,optional" json:"NotesAndComments"`
						}
					}
				}
			}
			FinancialTransaction        []FT1 `hl7:"FT1,optional" json:"FinancialTransaction"`
			ClinicalTrialIdentification []CTI `hl7:"CTI,optional" json:"ClinicalTrialIdentification"`
			Billing                     BLG   `hl7:"BLG,optional" json:"Billing"`
		}
	}
}
