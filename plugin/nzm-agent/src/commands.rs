use crate::ipc::{Request, Response, SendKeysParams, PaneIdParam};
use crate::state::State;
use serde::{Deserialize, Serialize};

/// DTO for pane information returned to CLI
#[derive(Debug, Serialize, Deserialize)]
pub struct PaneDto {
    pub id: u32,
    pub title: String,
    pub is_focused: bool,
    pub is_floating: bool,
}

/// Dispatch a request to the appropriate handler
pub fn dispatch_command(req: &Request, state: &State) -> Response {
    match req.action.as_str() {
        "list_panes" => handle_list_panes(req, state),
        "get_pane_info" => handle_get_pane_info(req, state),
        "send_keys" => handle_send_keys_validate(req, state),
        "send_interrupt" => handle_send_interrupt_validate(req, state),
        _ => Response {
            id: req.id.clone(),
            success: false,
            data: None,
            error: Some(format!("unknown action: {}", req.action)),
        },
    }
}

/// Handle list_panes action
fn handle_list_panes(req: &Request, state: &State) -> Response {
    let panes: Vec<PaneDto> = state.panes().iter().map(|p| PaneDto {
        id: p.id,
        title: p.title.clone(),
        is_focused: p.is_focused,
        is_floating: p.is_floating,
    }).collect();

    Response {
        id: req.id.clone(),
        success: true,
        data: Some(serde_json::json!({ "panes": panes })),
        error: None,
    }
}

/// Handle get_pane_info action
fn handle_get_pane_info(req: &Request, state: &State) -> Response {
    let params: Result<PaneIdParam, _> = serde_json::from_value(req.params.clone());

    match params {
        Ok(p) => {
            match state.get_pane(p.pane_id) {
                Some(pane) => Response {
                    id: req.id.clone(),
                    success: true,
                    data: Some(serde_json::json!({
                        "pane": PaneDto {
                            id: pane.id,
                            title: pane.title.clone(),
                            is_focused: pane.is_focused,
                            is_floating: pane.is_floating,
                        }
                    })),
                    error: None,
                },
                None => Response {
                    id: req.id.clone(),
                    success: false,
                    data: None,
                    error: Some(format!("pane not found: {}", p.pane_id)),
                },
            }
        }
        Err(e) => Response {
            id: req.id.clone(),
            success: false,
            data: None,
            error: Some(format!("invalid params: {}", e)),
        },
    }
}

/// Validate send_keys params (actual sending happens in lib.rs with Zellij API)
fn handle_send_keys_validate(req: &Request, state: &State) -> Response {
    let params: Result<SendKeysParams, _> = serde_json::from_value(req.params.clone());

    match params {
        Ok(p) => {
            // Verify pane exists
            if state.get_pane(p.pane_id).is_none() {
                return Response {
                    id: req.id.clone(),
                    success: false,
                    data: None,
                    error: Some(format!("pane not found: {}", p.pane_id)),
                };
            }

            // Return success with params for lib.rs to execute
            Response {
                id: req.id.clone(),
                success: true,
                data: Some(serde_json::json!({
                    "action": "send_keys",
                    "pane_id": p.pane_id,
                    "text": p.text,
                    "enter": p.enter,
                })),
                error: None,
            }
        }
        Err(e) => Response {
            id: req.id.clone(),
            success: false,
            data: None,
            error: Some(format!("invalid params: {}", e)),
        },
    }
}

/// Validate send_interrupt params
fn handle_send_interrupt_validate(req: &Request, state: &State) -> Response {
    let params: Result<PaneIdParam, _> = serde_json::from_value(req.params.clone());

    match params {
        Ok(p) => {
            if state.get_pane(p.pane_id).is_none() {
                return Response {
                    id: req.id.clone(),
                    success: false,
                    data: None,
                    error: Some(format!("pane not found: {}", p.pane_id)),
                };
            }

            Response {
                id: req.id.clone(),
                success: true,
                data: Some(serde_json::json!({
                    "action": "send_interrupt",
                    "pane_id": p.pane_id,
                })),
                error: None,
            }
        }
        Err(e) => Response {
            id: req.id.clone(),
            success: false,
            data: None,
            error: Some(format!("invalid params: {}", e)),
        },
    }
}

/// Validate send_keys parameters
#[allow(dead_code)]
pub fn validate_send_keys_params(_params: &SendKeysParams) -> Result<(), String> {
    // pane_id validation happens when we look it up
    // text can be empty (might just press enter)
    Ok(())
}

#[cfg(test)]
mod tests {
    use super::*;
    use zellij_tile::prelude::{PaneInfo, PaneManifest};

    fn create_test_pane(id: u32, title: &str, is_plugin: bool) -> PaneInfo {
        PaneInfo {
            id,
            is_plugin,
            title: title.to_string(),
            is_focused: id == 1, // First pane is focused
            is_fullscreen: false,
            is_floating: false,
            is_suppressed: false,
            ..Default::default()
        }
    }

    fn create_manifest_with_panes(panes: Vec<PaneInfo>) -> PaneManifest {
        let mut manifest = PaneManifest::default();
        manifest.panes.insert(0, panes);
        manifest
    }

    fn create_test_state() -> State {
        let mut state = State::default();
        state.update_panes(create_manifest_with_panes(vec![
            create_test_pane(1, "proj__cc_1", false),
            create_test_pane(2, "proj__cc_2", false),
        ]));
        state
    }

    #[test]
    fn test_handle_list_panes_returns_pane_array() {
        let state = create_test_state();
        let req = Request {
            id: "123".to_string(),
            action: "list_panes".to_string(),
            params: serde_json::Value::Null,
        };

        let result = dispatch_command(&req, &state);

        assert!(result.success);
        assert_eq!(result.id, "123");
        let data = result.data.unwrap();
        let panes: Vec<PaneDto> = serde_json::from_value(data["panes"].clone()).unwrap();
        assert_eq!(panes.len(), 2);
        assert_eq!(panes[0].id, 1);
        assert_eq!(panes[0].title, "proj__cc_1");
        assert!(panes[0].is_focused);
    }

    #[test]
    fn test_handle_list_panes_empty_state() {
        let state = State::default();
        let req = Request {
            id: "1".to_string(),
            action: "list_panes".to_string(),
            params: serde_json::Value::Null,
        };

        let result = dispatch_command(&req, &state);

        assert!(result.success);
        let data = result.data.unwrap();
        let panes: Vec<PaneDto> = serde_json::from_value(data["panes"].clone()).unwrap();
        assert!(panes.is_empty());
    }

    #[test]
    fn test_handle_get_pane_info_found() {
        let state = create_test_state();
        let req = Request {
            id: "1".to_string(),
            action: "get_pane_info".to_string(),
            params: serde_json::json!({"pane_id": 1}),
        };

        let result = dispatch_command(&req, &state);

        assert!(result.success);
        let data = result.data.unwrap();
        let pane: PaneDto = serde_json::from_value(data["pane"].clone()).unwrap();
        assert_eq!(pane.id, 1);
        assert_eq!(pane.title, "proj__cc_1");
    }

    #[test]
    fn test_handle_get_pane_info_not_found() {
        let state = create_test_state();
        let req = Request {
            id: "1".to_string(),
            action: "get_pane_info".to_string(),
            params: serde_json::json!({"pane_id": 999}),
        };

        let result = dispatch_command(&req, &state);

        assert!(!result.success);
        assert!(result.error.unwrap().contains("pane not found"));
    }

    #[test]
    fn test_handle_send_keys_valid() {
        let state = create_test_state();
        let req = Request {
            id: "1".to_string(),
            action: "send_keys".to_string(),
            params: serde_json::json!({
                "pane_id": 1,
                "text": "hello",
                "enter": true
            }),
        };

        let result = dispatch_command(&req, &state);

        assert!(result.success);
        let data = result.data.unwrap();
        assert_eq!(data["pane_id"], 1);
        assert_eq!(data["text"], "hello");
        assert_eq!(data["enter"], true);
    }

    #[test]
    fn test_handle_send_keys_pane_not_found() {
        let state = create_test_state();
        let req = Request {
            id: "1".to_string(),
            action: "send_keys".to_string(),
            params: serde_json::json!({
                "pane_id": 999,
                "text": "hello",
                "enter": false
            }),
        };

        let result = dispatch_command(&req, &state);

        assert!(!result.success);
        assert!(result.error.unwrap().contains("pane not found"));
    }

    #[test]
    fn test_handle_send_keys_invalid_params() {
        let state = create_test_state();
        let req = Request {
            id: "1".to_string(),
            action: "send_keys".to_string(),
            params: serde_json::json!({"wrong_field": 123}),
        };

        let result = dispatch_command(&req, &state);

        assert!(!result.success);
        assert!(result.error.unwrap().contains("invalid params"));
    }

    #[test]
    fn test_handle_send_interrupt_valid() {
        let state = create_test_state();
        let req = Request {
            id: "1".to_string(),
            action: "send_interrupt".to_string(),
            params: serde_json::json!({"pane_id": 1}),
        };

        let result = dispatch_command(&req, &state);

        assert!(result.success);
        let data = result.data.unwrap();
        assert_eq!(data["action"], "send_interrupt");
        assert_eq!(data["pane_id"], 1);
    }

    #[test]
    fn test_handle_unknown_action() {
        let state = State::default();
        let req = Request {
            id: "1".to_string(),
            action: "unknown_action".to_string(),
            params: serde_json::Value::Null,
        };

        let result = dispatch_command(&req, &state);

        assert!(!result.success);
        assert!(result.error.unwrap().contains("unknown action"));
    }

    #[test]
    fn test_validate_send_keys_params_valid() {
        let params = SendKeysParams {
            pane_id: 1,
            text: "hello".to_string(),
            enter: true,
        };
        assert!(validate_send_keys_params(&params).is_ok());
    }

    #[test]
    fn test_validate_send_keys_params_empty_text() {
        let params = SendKeysParams {
            pane_id: 1,
            text: "".to_string(),
            enter: false,
        };
        // Empty text is allowed (might just press enter)
        assert!(validate_send_keys_params(&params).is_ok());
    }
}
