package banner

import (
	cli "github.com/velocitykode/velocity-cli"
)

// SmallSimple returns a smaller ASCII banner
func SmallSimple() string {
	return cli.StylePrimary(`
██╗   ██╗███████╗██╗      ██████╗  ██████╗██╗████████╗██╗   ██╗     ██████╗██╗     ██╗
██║   ██║██╔════╝██║     ██╔═══██╗██╔════╝██║╚══██╔══╝╚██╗ ██╔╝    ██╔════╝██║     ██║
██║   ██║█████╗  ██║     ██║   ██║██║     ██║   ██║    ╚████╔╝     ██║     ██║     ██║
╚██╗ ██╔╝██╔══╝  ██║     ██║   ██║██║     ██║   ██║     ╚██╔╝      ██║     ██║     ██║
 ╚████╔╝ ███████╗███████╗╚██████╔╝╚██████╗██║   ██║      ██║       ╚██████╗███████╗██║
  ╚═══╝  ╚══════╝╚══════╝ ╚═════╝  ╚═════╝╚═╝   ╚═╝      ╚═╝        ╚═════╝╚══════╝╚═╝`)
}

// Small returns a compact ASCII banner
func Small() string {
	return cli.StylePrimary(`
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
	return cli.StylePrimary(`
╔════════════════════════════════════╗
║     VELOCITY CLI                   ║
║     Web Framework for Go           ║
╚════════════════════════════════════╝`)
}

// Compact returns a very compact banner
func Compact() string {
	return cli.StylePrimary(`
 VELOCITY CLI 
══════════════════════════════════════════════════
 The Go Web Framework for Rapid Development`)
}
