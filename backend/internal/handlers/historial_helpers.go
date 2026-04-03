package handlers

import "strings"

func normalizeHistorialTipoOperacion(s string) string {
	s = strings.TrimSpace(strings.ToLower(s))
	switch s {
	case "egreso":
		return "egreso"
	case "mantenimiento":
		return "mantenimiento"
	default:
		return "ingreso"
	}
}
