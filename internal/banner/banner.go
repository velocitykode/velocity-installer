package banner

import (
	"github.com/velocitykode/velocity-installer/internal/colors"
)

// Simple returns a simple ASCII banner
func Simple() string {
	style := colors.BrandStyle
	return style.Render(`
██╗   ██╗███████╗██╗      ██████╗  ██████╗██╗████████╗██╗   ██╗
██║   ██║██╔════╝██║     ██╔═══██╗██╔════╝██║╚══██╔══╝╚██╗ ██╔╝
██║   ██║█████╗  ██║     ██║   ██║██║     ██║   ██║    ╚████╔╝ 
╚██╗ ██╔╝██╔══╝  ██║     ██║   ██║██║     ██║   ██║     ╚██╔╝  
 ╚████╔╝ ███████╗███████╗╚██████╔╝╚██████╗██║   ██║      ██║   
  ╚═══╝  ╚══════╝╚══════╝ ╚═════╝  ╚═════╝╚═╝   ╚═╝      ╚═╝   
`)
}

// Block returns a blocky ASCII banner
func Block() string {
	primary := colors.BrandStyle
	accent := colors.BrandHoverStyle

	banner := primary.Render(`
█▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀█
█                                                        █
█  ██    ██ ███████ ██       ██████   ██████ ██ ████████▄█
█  ██    ██ ██      ██      ██    ██ ██      ██    ██    █
█  ██    ██ █████   ██      ██    ██ ██      ██    ██    █
█   ██  ██  ██      ██      ██    ██ ██      ██    ██    █
█    ████   ███████ ███████  ██████   ██████ ██    ██    █
█                                                        █
█▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄█
`) + accent.Render(`
      The Go Web Framework for Rapid Development`)

	return banner
}

// CompactBox returns a compact boxed ASCII banner
func CompactBox() string {
	style := colors.BrandStyle

	return style.Render(`
╔══════════════════════════════════════╗
║  ▌ ▌▛▀▌▌  ▞▀▖▞▀▖▀▛▘▀▛▘▌ ▌          ║
║  ▚▞ ▙▄ ▌  ▌ ▌▌  ▌▌  ▌ ▝▞           ║
║  ▝▘ ▌  ▀▀▘▝▀ ▝▀ ▀▀  ▀  ▀           ║
╚══════════════════════════════════════╝`)
}

// Retro returns a retro-style ASCII banner
func Retro() string {
	style := colors.BrandStyle
	return style.Render(`
 ▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄▄
 ██                                                         ██
 ██  ██    ██ ████████ ██       ███████  ███████ ████ ███████
 ██  ██    ██ ██       ██      ██    ██ ██       ██     ██  ██
 ██  ██    ██ ████████ ██      ██    ██ ██       ██     ██  ██
 ██   ██  ██  ██       ██      ██    ██ ██       ██     ██  ██
 ██    ████   ████████ ████████ ███████  ███████ ████   ██  ██
 ██                                                         ██
 ▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀▀`)
}

// Shadow returns a shadow-style ASCII banner
func Shadow() string {
	primary := colors.BrandStyle
	shadow := colors.MutedStyle

	return primary.Render(`
██╗   ██╗███████╗██╗      ██████╗  ██████╗██╗████████╗██╗   ██╗
██║   ██║██╔════╝██║     ██╔═══██╗██╔════╝██║╚══██╔══╝╚██╗ ██╔╝
██║   ██║█████╗  ██║     ██║   ██║██║     ██║   ██║    ╚████╔╝ 
╚██╗ ██╔╝██╔══╝  ██║     ██║   ██║██║     ██║   ██║     ╚██╔╝  
 ╚████╔╝ ███████╗███████╗╚██████╔╝╚██████╗██║   ██║      ██║   `) + "\n" +
		shadow.Render(` ╚═══╝  ╚══════╝╚══════╝ ╚═════╝  ╚═════╝╚═╝   ╚═╝      ╚═╝   `)
}
