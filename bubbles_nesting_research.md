# Bubbletea/Bubbles Nesting Patterns Research

## Overview

This document analyzes patterns for implementing hierarchical/nested displays in Bubbletea and Bubbles components, specifically focused on implementing Emmy's agent hierarchy in the sidebar.

## Key Findings

### 1. Lipgloss Tree Component (Best Match for Emmy)

The lipgloss tree component (`resources/lipgloss/tree/`) provides a complete solution for rendering hierarchical tree structures. This is the most suitable option for Emmy's nested agent display.

#### Key Features:
- **Built-in Tree Rendering**: Handles all the visual hierarchy with proper Unicode box-drawing characters
- **Flexible Node Types**: Supports both leaf nodes and tree nodes (with children)
- **Customizable Styling**: Per-node styling with `StyleFunc` callbacks
- **Hidden Nodes**: Built-in support for hiding/showing nodes
- **Multiple Enumerators**: Different styles for tree branches (default, rounded)

#### Example Usage:
```go
t := tree.New().
    Child(
        "Orchestrator",
        tree.Root("UI Builder").
            Child(
                "Component A",
                tree.Root("Sub-component").
                    Child("Handler"),
                "Component B",
            ),
        tree.Root("Backend").
            Child("API Handler", "Database"),
    )
```

#### Rendering Output:
```
├── Orchestrator
├── UI Builder
│   ├── Component A
│   ├── Sub-component
│   │   └── Handler
│   └── Component B
└── Backend
    ├── API Handler
    └── Database
```

### 2. Bubbles List Component

The bubbles list component (`resources/bubbles/list/`) is primarily designed for flat lists with filtering and pagination. While it doesn't directly support nesting, it provides:

- **Item Interface**: Could be extended to include hierarchy metadata
- **Custom Rendering**: ItemDelegate allows custom rendering per item
- **Keyboard Navigation**: Built-in up/down navigation
- **Filtering**: Fuzzy search capabilities

#### Limitations for Nesting:
- No built-in tree rendering
- Would require manual indentation handling
- No expand/collapse functionality out of the box

### 3. Existing Emmy Sidebar Implementation

The current Emmy sidebar (`internal/ui/sidebar/list.go`) implements:
- Basic flat list of agents
- Keyboard navigation (up/down/enter)
- Focus management
- Visual selection indicator
- No hierarchy support

### 4. Patterns from Other Examples

#### Superfile Sidebar
- Uses a flat list with filtering
- No hierarchical display
- Focus on file navigation

#### Claude Squad
- Flat list of instances
- No nesting patterns

## Recommended Implementation Approach

### Option 1: Lipgloss Tree Integration (Recommended)

1. **Replace Current List with Tree Structure**:
   ```go
   type Agent struct {
       ID       string
       Name     string
       Status   string
       Active   bool
       Parent   string     // Parent agent ID
       Children []string   // Child agent IDs
   }
   ```

2. **Build Tree from Agent Data**:
   ```go
   func buildAgentTree(agents []Agent) *tree.Tree {
       t := tree.New()
       
       // Build hierarchy from flat list
       agentMap := make(map[string]*Agent)
       for i := range agents {
           agentMap[agents[i].ID] = &agents[i]
       }
       
       // Add root agents and recursively add children
       for _, agent := range agents {
           if agent.Parent == "" {
               addAgentToTree(t, &agent, agentMap)
           }
       }
       
       return t
   }
   ```

3. **Maintain Current Navigation**:
   - Track current position in flattened tree
   - Map cursor position to tree nodes
   - Support expand/collapse with left/right keys

### Option 2: Extend Current List Component

1. **Add Hierarchy Metadata**:
   ```go
   type Agent struct {
       // ... existing fields
       Level    int    // Nesting level
       Expanded bool   // For collapsible nodes
       HasChildren bool
   }
   ```

2. **Custom Rendering with Indentation**:
   ```go
   func renderAgent(agent Agent) string {
       indent := strings.Repeat("  ", agent.Level)
       icon := "├── "
       if agent.HasChildren {
           if agent.Expanded {
               icon = "▼ "
           } else {
               icon = "▶ "
           }
       }
       return indent + icon + agent.Name
   }
   ```

## Visual Indicators for Hierarchy

### Tree Branch Styles:
- **Default**: `├──`, `└──`, `│`
- **Rounded**: `├──`, `╰──`, `│`
- **ASCII**: `+--`, `+--`, `|`

### Expansion Indicators:
- **Expanded**: `▼`, `▾`, `⊟`, `-`
- **Collapsed**: `▶`, `▸`, `⊞`, `+`
- **Leaf**: `•`, `◦`, `-`

## Keyboard Navigation Patterns

### Standard Tree Navigation:
- **Up/Down**: Move between visible items
- **Left**: Collapse current node or move to parent
- **Right**: Expand current node or move to first child
- **Enter**: Select/activate current node
- **Space**: Toggle expand/collapse

### Advanced Navigation:
- **Home/End**: Jump to first/last visible item
- **Page Up/Down**: Scroll by page
- **Ctrl+Up/Down**: Move to previous/next sibling at same level

## State Management for Expanded/Collapsed Nodes

```go
type TreeState struct {
    expanded map[string]bool  // Track which nodes are expanded
    cursor   int             // Current cursor position
    visible  []string         // Flattened list of visible node IDs
}

func (s *TreeState) toggleNode(nodeID string) {
    s.expanded[nodeID] = !s.expanded[nodeID]
    s.rebuildVisible() // Rebuild visible node list
}
```

## Challenges and Solutions

### Challenge 1: Dynamic Updates
- **Problem**: Tree structure changes when agents spawn/terminate
- **Solution**: Rebuild tree on agent list changes, preserve expansion state

### Challenge 2: Large Hierarchies
- **Problem**: Deep nesting might exceed sidebar width
- **Solution**: Limit visible depth, use horizontal scrolling, or abbreviate deep paths

### Challenge 3: Performance
- **Problem**: Rendering large trees might be slow
- **Solution**: Only render visible nodes, use virtual scrolling

## Example Implementation for Emmy

```go
// Enhanced sidebar model
type Model struct {
    tree     *tree.Tree
    agents   []Agent
    cursor   int
    expanded map[string]bool
    // ... other fields
}

// Update tree rendering
func (m Model) View() string {
    // Build tree with custom styling
    t := tree.New().
        EnumeratorStyle(m.styles.TreeBranch).
        ItemStyleFunc(func(children tree.Children, i int) lipgloss.Style {
            // Highlight selected item
            if i == m.cursor {
                return m.styles.SelectedItem
            }
            return m.styles.Item
        })
    
    // Add agents to tree
    m.buildAgentTree(t)
    
    return m.styles.Border.Render(t.String())
}
```

## Conclusion

The lipgloss tree component provides the most complete solution for implementing Emmy's hierarchical agent display. It handles all the complex rendering logic while allowing customization through styles and enumerators. The main implementation work would be:

1. Converting the flat agent list to a tree structure
2. Mapping keyboard navigation to tree traversal
3. Managing expand/collapse state
4. Integrating with Emmy's existing styling system

This approach would provide a professional, visually appealing nested display that matches Emmy's design aesthetic while maintaining the current interaction patterns.