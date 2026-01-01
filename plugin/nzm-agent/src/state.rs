use std::collections::HashMap;
use zellij_tile::prelude::{PaneInfo, PaneManifest};

/// Tracks the current state of panes in the Zellij session
#[derive(Default)]
pub struct State {
    panes: Vec<PaneInfo>,
    pane_by_id: HashMap<u32, usize>,
}

impl State {
    /// Update pane state from a PaneManifest event
    pub fn update_panes(&mut self, manifest: PaneManifest) {
        self.panes.clear();
        self.pane_by_id.clear();

        for (_tab_idx, tab_panes) in manifest.panes {
            for pane in tab_panes {
                // Only track terminal panes, not plugin panes
                if !pane.is_plugin {
                    let idx = self.panes.len();
                    self.pane_by_id.insert(pane.id, idx);
                    self.panes.push(pane);
                }
            }
        }
    }

    /// Get all tracked panes
    pub fn panes(&self) -> &[PaneInfo] {
        &self.panes
    }

    /// Get a pane by its ID
    pub fn get_pane(&self, id: u32) -> Option<&PaneInfo> {
        self.pane_by_id.get(&id).map(|&idx| &self.panes[idx])
    }

    /// Get a pane by its title
    pub fn get_pane_by_title(&self, title: &str) -> Option<&PaneInfo> {
        self.panes.iter().find(|p| p.title == title)
    }

    /// Get panes matching a title pattern (prefix match)
    pub fn get_panes_by_prefix(&self, prefix: &str) -> Vec<&PaneInfo> {
        self.panes.iter().filter(|p| p.title.starts_with(prefix)).collect()
    }
}

#[cfg(test)]
mod tests {
    use super::*;

    // Helper to create a test PaneInfo
    fn create_test_pane(id: u32, title: &str, is_plugin: bool) -> PaneInfo {
        PaneInfo {
            id,
            is_plugin,
            title: title.to_string(),
            is_focused: false,
            is_fullscreen: false,
            is_floating: false,
            is_suppressed: false,
            ..Default::default()
        }
    }

    // Helper to create a PaneManifest with given panes
    fn create_manifest_with_panes(panes: Vec<PaneInfo>) -> PaneManifest {
        let mut manifest = PaneManifest::default();
        manifest.panes.insert(0, panes); // Tab 0
        manifest
    }

    #[test]
    fn test_empty_state_has_no_panes() {
        let state = State::default();
        assert!(state.panes().is_empty());
    }

    #[test]
    fn test_update_panes_stores_terminal_panes() {
        let mut state = State::default();

        let pane_info = create_test_pane(1, "test__cc_1", false);
        let manifest = create_manifest_with_panes(vec![pane_info]);

        state.update_panes(manifest);

        assert_eq!(state.panes().len(), 1);
        assert_eq!(state.panes()[0].id, 1);
        assert_eq!(state.panes()[0].title, "test__cc_1");
    }

    #[test]
    fn test_update_panes_excludes_plugin_panes() {
        let mut state = State::default();

        let terminal = create_test_pane(1, "test__cc_1", false);
        let plugin = create_test_pane(2, "nzm-agent", true);
        let manifest = create_manifest_with_panes(vec![terminal, plugin]);

        state.update_panes(manifest);

        assert_eq!(state.panes().len(), 1); // Only terminal pane
        assert_eq!(state.panes()[0].title, "test__cc_1");
    }

    #[test]
    fn test_update_panes_clears_previous_state() {
        let mut state = State::default();

        // First update
        state.update_panes(create_manifest_with_panes(vec![
            create_test_pane(1, "old_pane", false),
        ]));
        assert_eq!(state.panes().len(), 1);

        // Second update replaces all
        state.update_panes(create_manifest_with_panes(vec![
            create_test_pane(2, "new_pane", false),
        ]));
        assert_eq!(state.panes().len(), 1);
        assert_eq!(state.panes()[0].id, 2);
        assert_eq!(state.panes()[0].title, "new_pane");
    }

    #[test]
    fn test_get_pane_by_id_returns_pane() {
        let mut state = State::default();
        let pane = create_test_pane(5, "test__cod_1", false);
        state.update_panes(create_manifest_with_panes(vec![pane]));

        let found = state.get_pane(5);
        assert!(found.is_some());
        assert_eq!(found.unwrap().id, 5);
        assert_eq!(found.unwrap().title, "test__cod_1");
    }

    #[test]
    fn test_get_pane_by_id_returns_none_for_missing() {
        let state = State::default();
        assert!(state.get_pane(999).is_none());
    }

    #[test]
    fn test_get_pane_by_title_returns_pane() {
        let mut state = State::default();
        let pane = create_test_pane(3, "myproject__gmi_1", false);
        state.update_panes(create_manifest_with_panes(vec![pane]));

        let found = state.get_pane_by_title("myproject__gmi_1");
        assert!(found.is_some());
        assert_eq!(found.unwrap().id, 3);
    }

    #[test]
    fn test_get_pane_by_title_returns_none_for_missing() {
        let state = State::default();
        assert!(state.get_pane_by_title("nonexistent").is_none());
    }

    #[test]
    fn test_get_panes_by_prefix() {
        let mut state = State::default();
        state.update_panes(create_manifest_with_panes(vec![
            create_test_pane(1, "myproject__cc_1", false),
            create_test_pane(2, "myproject__cc_2", false),
            create_test_pane(3, "myproject__cod_1", false),
            create_test_pane(4, "other__cc_1", false),
        ]));

        let cc_panes = state.get_panes_by_prefix("myproject__cc_");
        assert_eq!(cc_panes.len(), 2);

        let myproject_panes = state.get_panes_by_prefix("myproject__");
        assert_eq!(myproject_panes.len(), 3);
    }

    #[test]
    fn test_multiple_tabs() {
        let mut state = State::default();
        let mut manifest = PaneManifest::default();
        manifest.panes.insert(0, vec![create_test_pane(1, "tab0_pane", false)]);
        manifest.panes.insert(1, vec![create_test_pane(2, "tab1_pane", false)]);

        state.update_panes(manifest);

        // Should have panes from both tabs
        assert_eq!(state.panes().len(), 2);
        assert!(state.get_pane(1).is_some());
        assert!(state.get_pane(2).is_some());
    }
}
