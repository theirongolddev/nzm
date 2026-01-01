mod ipc;
mod state;
mod commands;

// Only include plugin code when building for WASM
#[cfg(target_arch = "wasm32")]
mod plugin;

// Re-export for external use
pub use ipc::{Request, Response, SendKeysParams, PaneIdParam};
pub use state::State;
pub use commands::{dispatch_command, PaneDto};

// Plugin entry point (WASM only)
#[cfg(target_arch = "wasm32")]
pub use plugin::NzmAgent;
