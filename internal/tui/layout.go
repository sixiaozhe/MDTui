package tui

import "mdtui/internal/document"

// Heights include borders; available includes the one-row gaps between panels.
// Preserve the existing equal split, then return space from capped panels to
// panels that can still grow. If every panel is capped, leave the surplus blank.
func panelHeights(panels []document.Panel, available int) []int {
	heights := make([]int, len(panels))
	if len(panels) == 0 {
		return heights
	}
	budget := max(0, available-(len(panels)-1))
	share := budget / len(panels)
	remaining := budget
	for i, panel := range panels {
		height := share
		if i == len(panels)-1 {
			height = budget - share*i
		}
		if panel.MaxHeight > 0 {
			height = min(height, panel.MaxHeight)
		}
		heights[i] = height
		remaining -= height
	}
	for remaining > 0 {
		var grow []int
		for i, panel := range panels {
			if panel.MaxHeight == 0 || heights[i] < panel.MaxHeight {
				grow = append(grow, i)
			}
		}
		if len(grow) == 0 {
			break
		}
		share = max(1, remaining/len(grow))
		for j := len(grow) - 1; j >= 0 && remaining > 0; j-- {
			i := grow[j]
			extra := min(share, remaining)
			if panels[i].MaxHeight > 0 {
				extra = min(extra, panels[i].MaxHeight-heights[i])
			}
			heights[i] += extra
			remaining -= extra
		}
	}
	return heights
}
