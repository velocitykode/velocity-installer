package banner

import (
	"github.com/velocitykode/prism"
)

// SmallSimple returns a smaller ASCII banner
func SmallSimple() string {
	return prism.StylePrimary(`
██╗   ██╗███████╗██╗      ██████╗  ██████╗██╗████████╗██╗   ██╗     ██████╗██╗     ██╗
██║   ██║██╔════╝██║     ██╔═══██╗██╔════╝██║╚══██╔══╝╚██╗ ██╔╝    ██╔════╝██║     ██║
██║   ██║█████╗  ██║     ██║   ██║██║     ██║   ██║    ╚████╔╝     ██║     ██║     ██║
╚██╗ ██╔╝██╔══╝  ██║     ██║   ██║██║     ██║   ██║     ╚██╔╝      ██║     ██║     ██║
 ╚████╔╝ ███████╗███████╗╚██████╔╝╚██████╗██║   ██║      ██║       ╚██████╗███████╗██║
  ╚═══╝  ╚══════╝╚══════╝ ╚═════╝  ╚═════╝╚═╝   ╚═╝      ╚═╝        ╚═════╝╚══════╝╚═╝`)
}

// Small returns a compact ASCII banner
func Small() string {
	return prism.StylePrimary(`
╔═══════════════════════════════════════════════════════════════════════╗
║  ██╗   ██╗███████╗██╗      ██████╗  ██████╗██╗████████╗██╗   ██╗    ║
║  ██║   ██║██╔════╝██║     ██╔═══██╗██╔════╝██║╚══██╔══╝╚██╗ ██╔╝    ║
║  ╚██╗ ██╔╝█████╗  ██║     ██║   ██║██║     ██║   ██║    ╚████╔╝     ║
║   ╚████╔╝ ███████╗███████╗╚██████╔╝╚██████╗██║   ██║      ██║  CLI  ║
║    ╚═══╝  ╚══════╝╚══════╝ ╚═════╝  ╚═════╝╚═╝   ╚═╝      ╚═╝       ║
╚═══════════════════════════════════════════════════════════════════════╝`)
}

// Minimal returns a minimal ASCII banner
func Minimal() string {
	return prism.StylePrimary(`
╔════════════════════════════════════╗
║     VELOCITY CLI                   ║
║     Web Framework for Go           ║
╚════════════════════════════════════╝`)
}

// Compact returns a very compact banner
func Compact() string {
	return prism.StylePrimary(`
 VELOCITY CLI 
══════════════════════════════════════════════════
 The Go Web Framework for Rapid Development`)
}
