use serde::{Deserialize, Serialize};
use serde_json::Value;

/// Request from CLI to plugin via zellij pipe
#[derive(Debug, Deserialize)]
pub struct Request {
    pub id: String,
    pub action: String,
    #[serde(default)]
    pub params: Value,
}

/// Response from plugin to CLI
#[derive(Debug, Serialize)]
pub struct Response {
    pub id: String,
    pub success: bool,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub data: Option<Value>,
    #[serde(skip_serializing_if = "Option::is_none")]
    pub error: Option<String>,
}

/// Parameters for send_keys action
#[derive(Debug, Deserialize)]
pub struct SendKeysParams {
    pub pane_id: u32,
    pub text: String,
    #[serde(default)]
    pub enter: bool,
}

/// Parameters for actions that target a single pane
#[derive(Debug, Deserialize)]
pub struct PaneIdParam {
    pub pane_id: u32,
}

#[cfg(test)]
mod tests {
    use super::*;

    #[test]
    fn test_parse_valid_request() {
        let json = r#"{"id":"123","action":"list_panes","params":{}}"#;
        let req: Request = serde_json::from_str(json).unwrap();

        assert_eq!(req.id, "123");
        assert_eq!(req.action, "list_panes");
    }

    #[test]
    fn test_parse_request_with_params() {
        let json = r#"{"id":"456","action":"send_keys","params":{"pane_id":3,"text":"hello","enter":true}}"#;
        let req: Request = serde_json::from_str(json).unwrap();

        assert_eq!(req.action, "send_keys");
        let params: SendKeysParams = serde_json::from_value(req.params).unwrap();
        assert_eq!(params.pane_id, 3);
        assert_eq!(params.text, "hello");
        assert!(params.enter);
    }

    #[test]
    fn test_parse_request_without_params() {
        let json = r#"{"id":"789","action":"list_panes"}"#;
        let req: Request = serde_json::from_str(json).unwrap();

        assert_eq!(req.id, "789");
        assert_eq!(req.action, "list_panes");
        assert!(req.params.is_null());
    }

    #[test]
    fn test_parse_invalid_json_returns_error() {
        let json = "not valid json";
        let result: Result<Request, _> = serde_json::from_str(json);
        assert!(result.is_err());
    }

    #[test]
    fn test_parse_missing_required_fields() {
        let json = r#"{"id":"123"}"#; // missing action
        let result: Result<Request, _> = serde_json::from_str(json);
        assert!(result.is_err());
    }

    #[test]
    fn test_serialize_success_response() {
        let resp = Response {
            id: "123".to_string(),
            success: true,
            data: Some(serde_json::json!({"panes": []})),
            error: None,
        };
        let json = serde_json::to_string(&resp).unwrap();

        assert!(json.contains(r#""success":true"#));
        assert!(json.contains(r#""id":"123""#));
        assert!(json.contains(r#""panes""#));
        assert!(!json.contains(r#""error""#)); // None should be skipped
    }

    #[test]
    fn test_serialize_error_response() {
        let resp = Response {
            id: "123".to_string(),
            success: false,
            data: None,
            error: Some("pane not found".to_string()),
        };
        let json = serde_json::to_string(&resp).unwrap();

        assert!(json.contains(r#""success":false"#));
        assert!(json.contains(r#""error":"pane not found""#));
        assert!(!json.contains(r#""data""#)); // None should be skipped
    }

    #[test]
    fn test_send_keys_params_defaults() {
        let json = r#"{"pane_id":1,"text":"test"}"#;
        let params: SendKeysParams = serde_json::from_str(json).unwrap();

        assert_eq!(params.pane_id, 1);
        assert_eq!(params.text, "test");
        assert!(!params.enter); // Default is false
    }

    #[test]
    fn test_pane_id_param() {
        let json = r#"{"pane_id":42}"#;
        let param: PaneIdParam = serde_json::from_str(json).unwrap();

        assert_eq!(param.pane_id, 42);
    }
}
