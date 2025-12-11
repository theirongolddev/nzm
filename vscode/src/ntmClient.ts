import * as cp from 'child_process';
import * as vscode from 'vscode';

export interface RobotStatus {
    generated_at: string;
    system: {
        version: string;
        os: string;
    };
    sessions: SessionInfo[];
    summary: StatusSummary;
    beads?: BeadsStatus;
    agent_mail?: AgentMailStatus;
}

export interface BeadsStatus {
    available: boolean;
    total: number;
    open: number;
    in_progress: number;
    blocked: number;
    ready: number;
    closed: number;
}

export interface AgentMailStatus {
    available: boolean;
    server_url?: string;
}

export interface SessionInfo {
    name: string;
    exists: boolean;
    attached: boolean;
    windows: number;
    panes: number;
    agents: AgentInfo[];
}

export interface AgentInfo {
    type: string;
    variant?: string;
    pane: string;
    is_active: boolean;
}

export interface StatusSummary {
    total_sessions: number;
    total_agents: number;
    attached_count: number;
}

export interface RobotMail {
    generated_at: string;
    project_key?: string;
    available: boolean;
    server_url?: string;
    agents?: MailAgent[];
    locks?: MailLock[];
    error?: string;
}

export interface MailAgent {
    name: string;
    program?: string;
    model?: string;
    unread_count?: number;
    urgent_count?: number;
    last_active_ts?: string;
}

export interface MailLock {
    id: number;
    path_pattern: string;
    agent_name: string;
    exclusive: boolean;
    reason?: string;
    expires_ts?: string;
    created_ts?: string;
    released_ts?: string | null;
}

export class NtmClient {
    private binaryPath: string;

    constructor() {
        const config = vscode.workspace.getConfiguration('ntm');
        this.binaryPath = config.get<string>('binaryPath', 'ntm');
    }

    private async run(args: string[]): Promise<string> {
        return new Promise((resolve, reject) => {
            cp.execFile(this.binaryPath, args, (err, stdout, stderr) => {
                if (err) {
                    const detail = stderr ? `: ${stderr}` : '';
                    reject(new Error(`ntm ${args.join(' ')} failed${detail}`));
                    return;
                }
                resolve(stdout);
            });
        });
    }

    async getStatus(): Promise<RobotStatus> {
        try {
            const stdout = await this.run(['--robot-status']);
            return JSON.parse(stdout);
        } catch (e) {
            throw new Error(`Failed to get status: ${e}`);
        }
    }

    async spawn(session: string, args: string[]): Promise<void> {
        await this.run(['spawn', session, ...args, '--json']);
    }

    async send(session: string, prompt: string, targets: string[] = []): Promise<void> {
        const args = ['send', session];
        
        if (targets.includes('all')) {
            args.push('--all');
        } else {
            if (targets.includes('cc')) args.push('--cc');
            if (targets.includes('cod')) args.push('--cod');
            if (targets.includes('gmi')) args.push('--gmi');
        }

        args.push(prompt);
        await this.run(args);
    }
    
    async checkAvailable(): Promise<boolean> {
        try {
            await this.run(['--version']);
            return true;
        } catch {
            return false;
        }
    }

    async getMail(): Promise<RobotMail> {
        try {
            const stdout = await this.run(['--robot-mail']);
            return JSON.parse(stdout);
        } catch (e) {
            throw new Error(`Failed to get mail state: ${e}`);
        }
    }
}
