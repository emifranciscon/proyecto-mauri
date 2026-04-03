package handlers

import "strings"

func normalizeHistorialTipoOperacion(s string) string {
	s = strings.TrimSpace(strings.ToLower(s))
	if s == "egreso" {
		return "egreso"
	}
	return "ingreso"
}
