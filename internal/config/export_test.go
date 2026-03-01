package config

// NewWithFS expose newWithFS aux tests externes (package config_test).
// Ce fichier est compilé uniquement lors des tests (suffixe _test.go),
// donc NewWithFS n'est jamais visible dans le binaire final.
var NewWithFS = newWithFS
