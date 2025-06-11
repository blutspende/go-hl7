package hl7

const (
	ANNOTATION_FIELDSEPARATOR = "fieldseparator"
	ANNOTATION_DELIMITER      = "delimiter" // annotation that triggers the delimiters in the scanner to be reset
	ANNOTATION_REQUIRED       = "require"   // field-annotation: by default all fields are optinal
	ANNOTATION_OPTIONAL       = "optional"  // record-annotation: by default all records are mandatory
	ANNOTATION_SEQUENCE       = "sequence"  // indicating that a sequence number should be generated (output only)
	ANNOTATION_LONGDATE       = "longdate"
)

type LineBreak int

const CR LineBreak = 0x13
const LF LineBreak = 0x10
const CRLF LineBreak = 0x1310

type FieldSeparator string

const StandardFieldSeparator FieldSeparator = "|"

/*
Notation  defines how the output format is build
ShortNotation will skip all delimiters to the right of the last value
StandardNotation will always produce as many delimiters as there are values in the export-format
*/
type Notation int

const StandardNotation = 1
const ShortNotation = 2
