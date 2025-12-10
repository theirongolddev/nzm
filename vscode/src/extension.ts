import * as vscode from 'vscode';
import { NtmClient } from './ntmClient';

export function activate(context: vscode.ExtensionContext) {
	const client = new NtmClient();
    const statusBarItem = vscode.window.createStatusBarItem(vscode.StatusBarAlignment.Left, 100);
    statusBarItem.command = 'ntm.showStatus';
    context.subscriptions.push(statusBarItem);

    // Initial check
    client.checkAvailable().then(available => {
        if (available) {
            updateStatus(client, statusBarItem);
            // Poll every 5s
            const interval = setInterval(() => updateStatus(client, statusBarItem), 5000);
            context.subscriptions.push({ dispose: () => clearInterval(interval) });
        } else {
            statusBarItem.text = '$(error) NTM missing';
            statusBarItem.show();
        }
    });

	let dispStatus = vscode.commands.registerCommand('ntm.showStatus', async () => {
		try {
            const status = await client.getStatus();
            const items = status.sessions.map(s => ({
                label: s.name,
                description: `${s.agents?.length || 0} agents`,
                detail: s.attached ? 'Attached' : 'Detached'
            }));
            
            vscode.window.showQuickPick(items, { placeHolder: 'Active NTM Sessions' });
        } catch (e) {
            vscode.window.showErrorMessage(`NTM Error: ${e}`);
        }
	});
    
    let dispSpawn = vscode.commands.registerCommand('ntm.spawn', async () => {
        const session = await vscode.window.showInputBox({ prompt: 'Session Name' });
        if (!session) return;
        try {
            await client.spawn(session, ['--cc=2']);
            vscode.window.showInformationMessage(`Spawned session ${session}`);
            updateStatus(client, statusBarItem);
        } catch (e) {
            vscode.window.showErrorMessage(`Spawn failed: ${e}`);
        }
    });

	context.subscriptions.push(dispStatus, dispSpawn);
}

async function updateStatus(client: NtmClient, item: vscode.StatusBarItem) {
    try {
        const status = await client.getStatus();
        const sessionCount = status.summary.total_sessions;
        const agentCount = status.summary.total_agents;
        
        if (sessionCount > 0) {
            item.text = `$(terminal) NTM: ${sessionCount} (${agentCount})`;
            item.tooltip = `${sessionCount} sessions, ${agentCount} agents`;
        } else {
            item.text = `$(terminal) NTM`;
            item.tooltip = "No active sessions";
        }
        item.show();
    } catch {
        item.text = `$(warning) NTM`;
        item.tooltip = "Error connecting to NTM";
        item.show();
    }
}

export function deactivate() {}