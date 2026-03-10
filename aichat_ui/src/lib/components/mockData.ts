import type { ChatMessage } from "./types";

export const demoMessages: ChatMessage[] = [
    {
        id: "demo-1",
        role: "user",
        content:
            "Can you explain how SvelteKit routing works and show me an example?",
    },
    {
        id: "demo-2",
        role: "assistant",
        content:
            "SvelteKit uses **filesystem-based routing**. Every file inside `src/routes` maps to a URL path.\n\n- `src/routes/+page.svelte` → `/`\n- `src/routes/about/+page.svelte` → `/about`\n- `src/routes/blog/[slug]/+page.svelte` → `/blog/:slug`\n\nYou can use **layout files** (`+layout.svelte`) to share UI between nested routes, and **server load functions** (`+page.server.ts`) to fetch data before rendering.",
        parts: [
            { type: "text", content: "" },
            {
                type: "code",
                content: `// src/routes/blog/[slug]/+page.server.ts\nimport type { PageServerLoad } from './$types';\nimport { error } from '@sveltejs/kit';\n\nexport const load: PageServerLoad = async ({ params }) => {\n  const post = await getPost(params.slug);\n\n  if (!post) {\n    throw error(404, 'Post not found');\n  }\n\n  return { post };\n};`,
                meta: { lang: "typescript", title: "+page.server.ts" },
            },
        ],
    },
    {
        id: "demo-3",
        role: "user",
        content: "Can you generate a project setup guide as an artifact?",
    },
    {
        id: "demo-4",
        role: "assistant",
        content: "I've prepared a project setup guide for you.",
        parts: [
            { type: "text", content: "" },
            {
                type: "artifact",
                meta: {
                    title: "SvelteKit Project Setup Guide",
                    description:
                        "Step-by-step instructions for bootstrapping a new SvelteKit application",
                },
                content: `## 1. Create the project\n\n\`\`\`bash\nnpx sv create my-app\ncd my-app\nnpm install\n\`\`\`\n\n \n\n## 2. Start development server\n\n\`\`\`bash\nnpm run dev\n\`\`\`\n\n\n\n## 3. Project structure\n\n| Path | Purpose |\n|------|---------|\n| \`/src/routes/\` | Pages and API routes |\n| \`/src/lib/\` | Shared components and utilities |\n| \`/static/\` | Static assets served as-is |`,
            },
        ],
    },
    {
        id: "demo-5",
        role: "user",
        content: "Deploy this to production on AWS.",
    },
    {
        id: "demo-6",
        role: "assistant",
        content:
            "I'll need your approval before I proceed with the production deployment.",
        parts: [
            { type: "text", content: "" },
            {
                type: "confirmation",
                meta: {
                    title: "Deploy to AWS Production",
                    description:
                        "This will build the application and deploy it to the us-east-1 production environment via CloudFormation.",
                    state: "approval-requested",
                    approval: { id: "deploy-001" },
                },
            },
        ],
    },
    {
        id: "demo-7",
        role: "user",
        content: "Show me the deployment plan.",
    },
    {
        id: "demo-8",
        role: "assistant",
        content: "Here's the deployment plan:",
        parts: [
            { type: "text", content: "" },
            {
                type: "plan",
                meta: {
                    title: "Production Deployment Plan",
                    description: "3 steps to deploy your SvelteKit app to AWS",
                    steps: [
                        "Run `npm run build` to create the production bundle",
                        "Upload build artifacts to S3 bucket `my-app-prod`",
                        "Invalidate CloudFront distribution cache",
                    ],
                },
            },
        ],
    },
    {
        id: "demo-9",
        role: "user",
        content: "Could you add a simple API endpoint example as well?",
    },
    {
        id: "demo-10",
        role: "assistant",
        content: "Certainly! Here is an example of an API endpoint in SvelteKit.",
        parts: [
            { type: "text", content: "" },
            {
                type: "code",
                content: `// src/routes/api/health/+server.ts\nimport { json } from '@sveltejs/kit';\nimport type { RequestHandler } from './$types';\n\nexport const GET: RequestHandler = async () => {\n    return json({\n        status: 'UP',\n        timestamp: new Date().toISOString()\n    });\n};`,
                meta: { lang: "typescript", title: "API Endpoint (GET)" },
            },
        ],
    },
    {
        id: "demo-11",
        role: "user",
        content: "Awesome, can you sketch out a database schema artifact for a Chat App?",
    },
    {
        id: "demo-12",
        role: "assistant",
        content: "I've created a database schema for your Chat Application.",
        parts: [
            { type: "text", content: "" },
            {
                type: "artifact",
                meta: {
                    title: "Chat App Database Schema",
                    description: "A simple relational schema for messages and users.",
                },
                content: `## Users Table\n\n| Column | Type | Constraints |\n|---|---|---|\n| id | UUID | PRIMARY KEY |\n| username | VARCHAR(50) | UNIQUE, NOT NULL |\n| created_at | TIMESTAMP | DEFAULT NOW() |\n\n## Messages Table\n\n| Column | Type | Constraints |\n|---|---|---|\n| id | UUID | PRIMARY KEY |\n| user_id | UUID | FOREIGN KEY (users.id) |\n| content | TEXT | NOT NULL |\n| sent_at | TIMESTAMP | DEFAULT NOW() |`,
            },
        ],
    }
];

export const suggestions = [
    {
        title: "Explain SvelteKit routing",
        text: "Can you explain how SvelteKit routing works and show me an example?",
    },
    {
        title: "Generate a setup guide",
        text: "Can you generate a project setup guide as an artifact?",
    },
    {
        title: "Deploy to production",
        text: "Deploy this to production on AWS.",
    },
    { title: "Show deployment plan", text: "Show me the deployment plan." },
];
