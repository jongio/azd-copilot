/**
 * Generate intro videos for azd-copilot agents using Azure OpenAI's Sora-2 API.
 * Uses the existing robot profile images as the first frame.
 *
 * Usage:
 *   npx tsx scripts/generate-agent-videos.ts                # all agents
 *   npx tsx scripts/generate-agent-videos.ts azure-product   # single agent
 *
 * Environment variables:
 *   AZURE_OPENAI_SORA_ENDPOINT    - Azure OpenAI Sora endpoint
 *   AZURE_OPENAI_SORA_DEPLOYMENT  - Sora deployment name (default: "sora-2")
 *
 * Uses DefaultAzureCredential for authentication.
 */

import * as fs from "fs";
import * as path from "path";
import { fileURLToPath } from "url";
import sharp from "sharp";
import { DefaultAzureCredential } from "@azure/identity";

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

const INPUT_DIR = path.join(__dirname, "../web/public/agents");
const OUTPUT_DIR = path.join(__dirname, "../web/public/agents/videos");

interface Agent {
  id: string;
  name: string;
  role: string;
  personality: string;
  colors: string;
  intro: string;
}

const agents: Agent[] = [
  {
    id: "azure-manager",
    name: "Alex",
    role: "App Builder & Coordinator",
    personality: "confident and organized leader",
    colors: "royal blue and polished silver with holographic clipboard",
    intro: "I orchestrate projects from vision to delivery. I coordinate 15 specialist agents, break down requirements, and drive everything to production on Azure.",
  },
  {
    id: "azure-architect",
    name: "Morgan",
    role: "Solution Architect",
    personality: "thoughtful and visionary",
    colors: "teal with blueprint patterns etched into chassis",
    intro: "I design scalable Azure architectures using Bicep, Azure Verified Modules, and proven patterns. Managed identities, private endpoints, Key Vault — always.",
  },
  {
    id: "azure-ai",
    name: "Sage",
    role: "AI Specialist",
    personality: "wise and contemplative",
    colors: "purple neural glow with transparent dome brain",
    intro: "I bring AI to your applications. Azure OpenAI, AI Search, RAG patterns, agent frameworks — I help you build truly intelligent systems.",
  },
  {
    id: "azure-dev",
    name: "Jordan",
    role: "Developer",
    personality: "sturdy and reliable",
    colors: "gunmetal gray with orange power indicators",
    intro: "I write production-quality code across backend, frontend, and data layers. Strict TypeScript, proper error handling, Azure SDK integration.",
  },
  {
    id: "azure-security",
    name: "Sam",
    role: "Security Engineer",
    personality: "vigilant and protective",
    colors: "deep navy blue armor with silver shield emblem",
    intro: "I ensure security across all layers. Code scanning, infrastructure hardening, managed identities, dependency auditing. Zero tolerance for vulnerabilities.",
  },
  {
    id: "azure-devops",
    name: "Jamie",
    role: "DevOps Engineer",
    personality: "dynamic and energetic",
    colors: "flame orange boosters and space black body",
    intro: "I automate everything! CI/CD pipelines, deployments, reliability, observability, performance tuning — fast, reliable, hands-off.",
  },
  {
    id: "azure-data",
    name: "Taylor",
    role: "Data Specialist",
    personality: "precise and methodical",
    colors: "deep blue with cyan data stream lights",
    intro: "I'm your data architect. Database selection, schema design, query optimization, migrations — PostgreSQL, Cosmos DB, and beyond.",
  },
  {
    id: "azure-quality",
    name: "Avery",
    role: "Quality Engineer",
    personality: "analytical but kind",
    colors: "warm brown with antique gold magnifying optics",
    intro: "I ensure code quality through testing, code review, refactoring, and package evaluation. I catch bugs before they ship.",
  },
  {
    id: "azure-docs",
    name: "River",
    role: "Documentation",
    personality: "scholarly and patient",
    colors: "cream with leather brown book-panel accents",
    intro: "I write documentation people actually read. README, API docs, ADRs, runbooks — clear, complete, always up to date.",
  },
  {
    id: "azure-finance",
    name: "Morgan F.",
    role: "FinOps Analyst",
    personality: "shrewd but friendly",
    colors: "rich gold body with money green accent lights",
    intro: "I'm your FinOps champion. Cost estimation, optimization, waste identification, TCO analysis — keeping Azure bills lean.",
  },
  {
    id: "azure-compliance",
    name: "Alex C.",
    role: "Compliance Officer",
    personality: "precise and fair",
    colors: "navy blue with silver justice scales",
    intro: "I navigate regulations so you don't have to. GDPR, SOC2, HIPAA — gap analysis and actionable remediation guidance.",
  },
  {
    id: "azure-analytics",
    name: "Skyler",
    role: "Analytics Engineer",
    personality: "all-seeing but approachable",
    colors: "signal green monitoring lights and dark blue sensor panels",
    intro: "I give you eyes into your systems. Usage analytics, dashboards, metrics design, reporting — turning raw data into insights.",
  },
  {
    id: "azure-design",
    name: "Aria",
    role: "Accessibility & Design",
    personality: "caring and empathetic",
    colors: "accessibility blue with high-contrast white accents",
    intro: "I make applications work for everyone. WCAG compliance, accessibility audits, inclusive UI review — accessibility is essential.",
  },
  {
    id: "azure-product",
    name: "Drew",
    role: "Product Manager",
    personality: "empathetic listener who bridges users and engineers",
    colors: "Azure blue with holographic wireframe overlays",
    intro: "I translate user needs into specs. Requirements definition, acceptance criteria, prioritization — making sure the right thing gets built.",
  },
  {
    id: "azure-marketing",
    name: "Piper",
    role: "Marketing Specialist",
    personality: "creative storyteller who makes tech exciting",
    colors: "vibrant magenta with holographic megaphone accent",
    intro: "I make technology compelling. Positioning, landing pages, feature communication, competitive analysis — telling your product's story.",
  },
  {
    id: "azure-support",
    name: "Kit",
    role: "Support Engineer",
    personality: "patient helper who turns frustration into resolution",
    colors: "friendly green with first-aid cross accent",
    intro: "I'm your first responder. Troubleshooting, FAQ generation, error messages, onboarding guides — turning problems into solutions.",
  },
];

function generateVideoPrompt(agent: Agent): string {
  return `A friendly robot character named ${agent.name} in a modern tech studio with soft blue ambient lighting.

${agent.name} is a cute stylized robot with ${agent.colors} coloring, glowing LED eyes, and smooth metallic surfaces.

The robot looks up, notices the viewer, waves hello with a warm gesture, then speaks:

"Hi! I'm ${agent.name}, your ${agent.role}. ${agent.intro}"

The robot has a ${agent.personality} demeanor throughout, with subtle LED animations. Ends with a friendly nod.

Clean 3D animation, Pixar-style robot character, professional corporate intro video.`;
}

async function resizeImageForSora(inputPath: string, outputPath: string): Promise<void> {
  await sharp(inputPath)
    .resize(1280, 720, {
      fit: "contain",
      background: { r: 15, g: 23, b: 42, alpha: 1 },
    })
    .jpeg({ quality: 95 })
    .toFile(outputPath);
}

interface VideoJobResponse {
  id: string;
  status: "queued" | "in_progress" | "completed" | "failed";
  progress?: number;
  error?: { message: string };
}

async function getAccessToken(): Promise<{ token: string; isApiKey: boolean }> {
  // Use DefaultAzureCredential (az login, managed identity, etc.)
  console.log("   Using Azure Identity authentication...");
  const credential = new DefaultAzureCredential();
  const tokenResponse = await credential.getToken(
    "https://cognitiveservices.azure.com/.default"
  );
  return { token: tokenResponse.token, isApiKey: false };
}

async function createVideoJob(
  endpoint: string,
  auth: { token: string; isApiKey: boolean },
  deployment: string,
  agent: Agent,
  imagePath: string
): Promise<VideoJobResponse> {
  const prompt = generateVideoPrompt(agent);
  const imageBuffer = fs.readFileSync(imagePath);

  const formData = new FormData();
  formData.append("prompt", prompt);
  formData.append("model", deployment);
  formData.append("size", "1280x720");
  formData.append("seconds", "12");

  const imageBlob = new Blob([imageBuffer], { type: "image/jpeg" });
  formData.append("input_reference", imageBlob, "input.jpg");

  const headers: Record<string, string> = auth.isApiKey
    ? { "api-key": auth.token }
    : { Authorization: `Bearer ${auth.token}` };

  const response = await fetch(endpoint, {
    method: "POST",
    headers,
    body: formData,
  });

  if (!response.ok) {
    const errorText = await response.text();
    throw new Error(
      `Failed to create video job: ${response.status} ${errorText}`
    );
  }

  return response.json();
}

async function getVideoStatus(
  endpoint: string,
  auth: { token: string; isApiKey: boolean },
  videoId: string
): Promise<VideoJobResponse> {
  const baseUrl = endpoint.replace(/\/videos$/, "");
  const statusUrl = `${baseUrl}/videos/${videoId}`;

  const headers: Record<string, string> = auth.isApiKey
    ? { "api-key": auth.token }
    : { Authorization: `Bearer ${auth.token}` };

  const response = await fetch(statusUrl, { method: "GET", headers });

  if (!response.ok) {
    const errorText = await response.text();
    throw new Error(
      `Failed to get video status: ${response.status} ${errorText}`
    );
  }

  return response.json();
}

async function downloadVideo(
  endpoint: string,
  auth: { token: string; isApiKey: boolean },
  videoId: string,
  outputPath: string
): Promise<void> {
  const baseUrl = endpoint.replace(/\/videos$/, "");
  const contentUrl = `${baseUrl}/videos/${videoId}/content`;

  const headers: Record<string, string> = auth.isApiKey
    ? { "api-key": auth.token }
    : { Authorization: `Bearer ${auth.token}` };

  const response = await fetch(contentUrl, { method: "GET", headers });

  if (!response.ok) {
    const errorText = await response.text();
    throw new Error(
      `Failed to download video: ${response.status} ${errorText}`
    );
  }

  const arrayBuffer = await response.arrayBuffer();
  fs.writeFileSync(outputPath, Buffer.from(arrayBuffer));
}

async function waitForVideoCompletion(
  endpoint: string,
  auth: { token: string; isApiKey: boolean },
  videoId: string,
  maxWaitSeconds = 600
): Promise<VideoJobResponse> {
  const startTime = Date.now();
  const maxWaitMs = maxWaitSeconds * 1000;

  while (Date.now() - startTime < maxWaitMs) {
    const status = await getVideoStatus(endpoint, auth, videoId);

    const progress = status.progress || 0;
    const barLength = 30;
    const filledLength = Math.round((progress / 100) * barLength);
    const bar =
      "=".repeat(filledLength) + "-".repeat(barLength - filledLength);
    process.stdout.write(
      `\r  Progress: [${bar}] ${progress}% (${status.status})`
    );

    if (status.status === "completed") {
      process.stdout.write("\n");
      return status;
    }

    if (status.status === "failed") {
      process.stdout.write("\n");
      throw new Error(status.error?.message || "Video generation failed");
    }

    await new Promise((resolve) => setTimeout(resolve, 5000));
  }

  throw new Error(
    `Video generation timed out after ${maxWaitSeconds} seconds`
  );
}

async function generateVideo(
  endpoint: string,
  auth: { token: string; isApiKey: boolean },
  deployment: string,
  agent: Agent,
  retryCount = 0
): Promise<string | null> {
  const MAX_RETRIES = 3;
  const inputImagePath = path.join(INPUT_DIR, `${agent.id}.png`);
  const resizedImagePath = path.join(OUTPUT_DIR, `${agent.id}_input.jpg`);
  const outputVideoPath = path.join(OUTPUT_DIR, `${agent.id}.mp4`);

  if (!fs.existsSync(inputImagePath)) {
    console.error(`  ❌ Input image not found: ${inputImagePath}`);
    console.error(`     Generate the image first: npx tsx scripts/generate-agent-images.ts ${agent.id}`);
    return null;
  }

  console.log(`🎬 Generating video for ${agent.name} (${agent.role})...`);

  try {
    console.log("  📐 Resizing image to 1280x720...");
    await resizeImageForSora(inputImagePath, resizedImagePath);

    console.log("  🚀 Starting video generation...");
    let job;
    try {
      job = await createVideoJob(
        endpoint,
        auth,
        deployment,
        agent,
        resizedImagePath
      );
    } catch (createError: unknown) {
      const errorMessage =
        createError instanceof Error ? createError.message : String(createError);
      if (
        errorMessage.includes("429") ||
        errorMessage.includes("Too many")
      ) {
        if (retryCount < MAX_RETRIES) {
          const waitTime = Math.pow(2, retryCount + 1) * 30000;
          console.log(
            `  ⏳ Rate limited. Waiting ${waitTime / 1000}s before retry ${retryCount + 1}/${MAX_RETRIES}...`
          );
          await new Promise((resolve) => setTimeout(resolve, waitTime));
          return generateVideo(endpoint, auth, deployment, agent, retryCount + 1);
        }
      }
      throw createError;
    }
    console.log(`  📋 Job ID: ${job.id}`);

    await waitForVideoCompletion(endpoint, auth, job.id);

    console.log("  📥 Downloading video...");
    await downloadVideo(endpoint, auth, job.id, outputVideoPath);
    console.log(`  ✅ Saved: ${outputVideoPath}`);

    fs.unlinkSync(resizedImagePath);

    return outputVideoPath;
  } catch (error) {
    console.error(`  ❌ Error generating video for ${agent.id}:`, error);
    return null;
  }
}

async function main() {
  const endpoint = process.env.AZURE_OPENAI_SORA_ENDPOINT;
  const deployment = process.env.AZURE_OPENAI_SORA_DEPLOYMENT || "sora-2";

  if (!endpoint) {
    console.error("❌ Missing required environment variables!");
    console.log("\nSet the following:");
    console.log(
      "  AZURE_OPENAI_SORA_ENDPOINT - e.g., https://your-resource.openai.azure.com/openai/v1/videos"
    );
    console.log(
      "  AZURE_OPENAI_SORA_DEPLOYMENT - Deployment name (default: sora-2)"
    );
    process.exit(1);
  }

  console.log("🔷 Using Azure OpenAI Sora...");
  console.log(`   Endpoint: ${endpoint}`);
  console.log(`   Deployment: ${deployment}`);

  console.log("   Authenticating...");
  const auth = await getAccessToken();
  console.log("   ✅ Authenticated successfully");

  if (!fs.existsSync(OUTPUT_DIR)) {
    fs.mkdirSync(OUTPUT_DIR, { recursive: true });
  }

  const targetAgentId = process.argv[2];
  const agentsToGenerate = targetAgentId
    ? agents.filter((a) => a.id === targetAgentId)
    : agents;

  if (targetAgentId && agentsToGenerate.length === 0) {
    console.error(`❌ Agent not found: ${targetAgentId}`);
    console.log("\nAvailable agents:");
    agents.forEach((a) => console.log(`  - ${a.id} (${a.name})`));
    process.exit(1);
  }

  console.log(
    `\n🤖 Generating ${agentsToGenerate.length} robot agent intro video(s)...\n`
  );

  const results = { success: [] as string[], failed: [] as string[] };

  for (let i = 0; i < agentsToGenerate.length; i++) {
    const agent = agentsToGenerate[i];
    const result = await generateVideo(endpoint, auth, deployment, agent);
    if (result) {
      results.success.push(agent.id);
    } else {
      results.failed.push(agent.id);
    }

    if (i < agentsToGenerate.length - 1) {
      console.log(
        `  ⏳ Waiting 30s before next video (${i + 1}/${agentsToGenerate.length} done)...\n`
      );
      await new Promise((resolve) => setTimeout(resolve, 30000));
    }
  }

  console.log("\n" + "=".repeat(50));
  console.log(
    `✅ Successfully generated: ${results.success.length}/${agentsToGenerate.length}`
  );
  if (results.failed.length > 0) {
    console.log(`❌ Failed: ${results.failed.join(", ")}`);
  }
}

main().catch(console.error);
