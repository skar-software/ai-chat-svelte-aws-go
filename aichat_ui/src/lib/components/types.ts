export type DemoPartType = "text" | "code" | "artifact" | "confirmation" | "plan" | "queue" | "citation";

export interface DemoPart {
    type: DemoPartType;
    content?: string;
    meta?: Record<string, any>;
}

export interface ChatMessage {
    id: string;
    role: "user" | "assistant";
    content: string;
    parts?: DemoPart[];
    /** Token usage (from backend usage event), optional */
    input_tokens?: number;
    output_tokens?: number;
}

/** Conversation summary for sidebar list (from GET /v1/conversations) */
export interface ConversationSummary {
    id: string;
    title?: string;
    created_at?: string;
    updated_at?: string;
}
