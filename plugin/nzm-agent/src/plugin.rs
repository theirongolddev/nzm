//! Zellij plugin entry point (WASM only)

use zellij_tile::prelude::*;
use crate::ipc::{Request, Response};
use crate::state::State;
use crate::commands;

#[derive(Default)]
pub struct NzmAgent {
    state: State,
    initialized: bool,
}

impl NzmAgent {
    pub fn is_initialized(&self) -> bool {
        self.initialized
    }
}

register_plugin!(NzmAgent);

impl ZellijPlugin for NzmAgent {
    fn load(&mut self, _config: std::collections::BTreeMap<String, String>) {
        request_permission(&[
            PermissionType::ReadApplicationState,
            PermissionType::WriteToStdin,
            PermissionType::RunCommands,
            PermissionType::MessageAndLaunchOtherPlugins,
        ]);
        subscribe(&[
            EventType::PaneUpdate,
            EventType::PermissionRequestResult,
        ]);
        self.initialized = true;
    }

    fn update(&mut self, event: Event) -> bool {
        match event {
            Event::PaneUpdate(manifest) => {
                self.state.update_panes(manifest);
                true
            }
            Event::PermissionRequestResult(result) => {
                if result == PermissionStatus::Granted {
                    // Permissions granted, we're ready
                }
                false
            }
            _ => false,
        }
    }

    fn pipe(&mut self, pipe_message: PipeMessage) -> bool {
        // Handle incoming IPC messages
        if let Some(payload) = pipe_message.payload {
            match serde_json::from_str::<Request>(&payload) {
                Ok(request) => {
                    let mut response = commands::dispatch_command(&request, &self.state);
                    response.id = request.id.clone();

                    // Execute actual Zellij commands if needed
                    if response.success {
                        if let Some(ref data) = response.data {
                            if let Some(action) = data.get("action").and_then(|v| v.as_str()) {
                                match action {
                                    "send_keys" => {
                                        if let (Some(pane_id), Some(text)) = (
                                            data.get("pane_id").and_then(|v| v.as_u64()),
                                            data.get("text").and_then(|v| v.as_str()),
                                        ) {
                                            let enter = data.get("enter").and_then(|v| v.as_bool()).unwrap_or(false);
                                            write_chars_to_pane_id(text, PaneId::Terminal(pane_id as u32));
                                            if enter {
                                                write_chars_to_pane_id("\n", PaneId::Terminal(pane_id as u32));
                                            }
                                        }
                                    }
                                    "send_interrupt" => {
                                        if let Some(pane_id) = data.get("pane_id").and_then(|v| v.as_u64()) {
                                            // Send Ctrl+C (ASCII 3)
                                            write_chars_to_pane_id("\x03", PaneId::Terminal(pane_id as u32));
                                        }
                                    }
                                    _ => {}
                                }
                            }
                        }
                    }

                    if let Ok(response_json) = serde_json::to_string(&response) {
                        // Send response back via CLI pipe
                        if let PipeSource::Cli(cli_id) = pipe_message.source {
                            cli_pipe_output(&cli_id.to_string(), &response_json);
                        }
                    }
                }
                Err(e) => {
                    let error_response = Response {
                        id: String::new(),
                        success: false,
                        data: None,
                        error: Some(format!("Failed to parse request: {}", e)),
                    };
                    if let Ok(response_json) = serde_json::to_string(&error_response) {
                        if let PipeSource::Cli(cli_id) = pipe_message.source {
                            cli_pipe_output(&cli_id.to_string(), &response_json);
                        }
                    }
                }
            }
        }
        false
    }

    fn render(&mut self, _rows: usize, _cols: usize) {
        // Plugin UI is minimal - just show status
        println!("NZM Agent | Panes: {}", self.state.panes().len());
    }
}
