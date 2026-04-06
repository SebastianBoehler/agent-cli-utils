//go:build darwin && cgo

// Force external linking on Darwin+cgo builds so the Mach-O includes LC_UUID.
// Newer macOS releases reject some internally linked test binaries that pull in
// Apple frameworks via the network stack.
package youtubecontrol

/*
 */
import "C"
