package banner

import (
	"github.com/velocitykode/velocity-installer/internal/colors"
)

// SmallSimple returns a smaller ASCII banner
func SmallSimple() string {
	return colors.BrandStyle.Render(`
██╗   ██╗███████╗██╗      ██████╗  ██████╗██╗████████╗██╗   ██╗     ██████╗██╗     ██╗
██║   ██║██╔════╝██║     ██╔═══██╗██╔════╝██║╚══██╔══╝╚██╗ ██╔╝    ██╔════╝██║     ██║
██║   ██║█████╗  ██║     ██║   ██║██║     ██║   ██║    ╚████╔╝     ██║     ██║     ██║
╚██╗ ██╔╝██╔══╝  ██║     ██║   ██║██║     ██║   ██║     ╚██╔╝      ██║     ██║     ██║
 ╚████╔╝ ███████╗███████╗╚██████╔╝╚██████╗██║   ██║      ██║       ╚██████╗███████╗██║
  ╚═══╝  ╚══════╝╚══════╝ ╚═════╝  ╚═════╝╚═╝   ╚═╝      ╚═╝        ╚═════╝╚══════╝╚═╝`)
}

// Small returns a compact ASCII banner
func Small() string {
	return colors.BrandStyle.Render(`
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
	return colors.BrandStyle.Render(`
╔════════════════════════════════════╗
║     VELOCITY CLI                   ║
║     Web Framework for Go           ║
╚════════════════════════════════════╝`)
}

// Compact returns a very compact banner
func Compact() string {
	return colors.BrandStyle.Render(`
 VELOCITY CLI 
══════════════════════════════════════════════════
 The Go Web Framework for Rapid Development`)
}
