/**
 * Generate AI robot images for all 16 azd-copilot agents using DALL-E 3.
 *
 * Usage:
 *   npx tsx scripts/generate-agent-images.ts              # all agents
 *   npx tsx scripts/generate-agent-images.ts azure-manager # single agent
 *
 * Environment variables (set via azd env or shell):
 *   AZURE_OPENAI_ENDPOINT          - Azure OpenAI endpoint
 *   AZURE_OPENAI_API_KEY           - Azure OpenAI API key
 *   AZURE_OPENAI_DALLE_DEPLOYMENT  - DALL-E deployment name (default: "dall-e-3")
 *   --- or ---
 *   OPENAI_API_KEY                 - Direct OpenAI API key
 */

import OpenAI, { AzureOpenAI } from "openai";
import * as fs from "fs";
import * as path from "path";
import { fileURLToPath } from "url";

const __filename = fileURLToPath(import.meta.url);
const __dirname = path.dirname(__filename);

const OUTPUT_DIR = path.join(__dirname, "../web/public/agents");

const agents = [
  {
    id: "azure-manager",
    name: "Alex",
    role: "App Builder & Coordinator",
    description: "A confident robot leader with a sleek royal blue metallic body. Has a warm LED smile, glowing blue eyes, and a small holographic clipboard floating nearby. Wears a subtle tie accent. Radiates calm authority.",
    colors: "royal blue and polished silver",
  },
  {
    id: "azure-architect",
    name: "Morgan",
    role: "Solution Architect",
    description: "A visionary robot with teal accents and blueprint patterns etched into its chassis. Has multiple optical sensors and holographic architecture diagrams floating around its head. Thoughtful, gazing upward.",
    colors: "teal and white with chrome details",
  },
  {
    id: "azure-ai",
    name: "Sage",
    role: "AI Specialist",
    description: "A wise robot with a transparent dome head showing a glowing neural network brain. Has ethereal purple lighting and contemplative pose. Mystical and intelligent aura.",
    colors: "purple with soft white inner glow",
  },
  {
    id: "azure-dev",
    name: "Jordan",
    role: "Developer",
    description: "A sturdy industrial robot with visible gears and cogs as design elements. Has tool-arm attachments and determined orange LED eyes. Built solid and dependable.",
    colors: "gunmetal gray with orange accents",
  },
  {
    id: "azure-security",
    name: "Sam",
    role: "Security Engineer",
    description: "A vigilant guardian robot with a shield emblem on its chest and scanning sensor eyes. Has reinforced armor plating but a friendly face. Alert and protective.",
    colors: "deep navy blue with silver shield",
  },
  {
    id: "azure-devops",
    name: "Jamie",
    role: "DevOps Engineer",
    description: "A dynamic rocket-themed robot with small boosters on its back and pipeline tubes visible. Has a helmet visor and excited, ready-for-launch expression.",
    colors: "flame orange and space black",
  },
  {
    id: "azure-data",
    name: "Taylor",
    role: "Data Specialist",
    description: "A precise robot with database cylinder elements in its torso. Has organized LED panels showing data streams and calm cyan glowing eyes. Methodical and organized appearance.",
    colors: "deep blue with cyan data lights",
  },
  {
    id: "azure-quality",
    name: "Avery",
    role: "Quality Engineer",
    description: "A detective-style robot with a magnifying glass monocle and observant glowing eyes. Has a deerstalker hat element and analytical but kind expression.",
    colors: "warm brown with antique gold accents",
  },
  {
    id: "azure-docs",
    name: "River",
    role: "Documentation",
    description: "A scholarly robot with book-shaped panels and a pen arm. Has reading glasses and a thoughtful, patient expression. Warm and helpful demeanor.",
    colors: "cream with leather brown accents",
  },
  {
    id: "azure-finance",
    name: "Morgan F.",
    role: "FinOps Analyst",
    description: "A savvy robot with coin-slot eyes and a piggy bank element on its body. Has a calculator arm and knowing smile. Friendly but financially smart.",
    colors: "gold with money green accents",
  },
  {
    id: "azure-compliance",
    name: "Alex C.",
    role: "Compliance Officer",
    description: "A precise robot with checklist patterns and balanced scales incorporated into design. Has a gavel arm accent and fair, measured expression.",
    colors: "navy blue with silver scales",
  },
  {
    id: "azure-analytics",
    name: "Skyler",
    role: "Analytics Engineer",
    description: "An all-seeing robot with satellite dish ears and multiple antenna. Has screens on its chest showing live metrics and graphs. Aware but friendly expression.",
    colors: "signal green with dark blue panels",
  },
  {
    id: "azure-design",
    name: "Aria",
    role: "Accessibility & Design",
    description: "An inclusive robot with universal design symbols and high contrast coloring. Has braille patterns on its body and caring, welcoming LED eyes.",
    colors: "accessibility blue with high contrast white",
  },
  {
    id: "azure-product",
    name: "Drew",
    role: "Product Manager",
    description: "A sleek robot with holographic wireframe overlays and a clipboard arm. Has empathetic blue eyes and an organized, thoughtful demeanor. Bridges creativity and logic.",
    colors: "Azure blue with holographic accents",
  },
  {
    id: "azure-marketing",
    name: "Piper",
    role: "Marketing Specialist",
    description: "A vibrant robot with a megaphone-shaped arm and glowing neon accents. Has an energetic pose with colorful holographic displays. Creative and expressive.",
    colors: "vibrant magenta with neon pink accents",
  },
  {
    id: "azure-support",
    name: "Kit",
    role: "Support Engineer",
    description: "A friendly first-responder robot with a first-aid cross on its chest and helpful extending arms. Has warm green eyes and a patient, reassuring expression.",
    colors: "friendly green with first-aid white cross",
  },
];

const BASE_PROMPT = `Create a friendly, stylized robot character portrait for a software development AI agent.

The robot should be:
- Cute and approachable with a distinct personality
- Clean digital illustration style with smooth surfaces
- Glowing LED eyes that convey emotion
- On a dark gradient tech background with subtle glow
- Portrait orientation, showing head and upper body
- Expressive design elements that match their role

This specific robot agent is:`;

async function generateImage(
  client: OpenAI | AzureOpenAI,
  agent: (typeof agents)[0],
  deploymentName?: string
): Promise<string | null> {
  const prompt = `${BASE_PROMPT}
Name: ${agent.name}
Role: ${agent.role}
Appearance: ${agent.description}
Color scheme: ${agent.colors}

Style: Clean digital illustration, professional portrait, friendly tech aesthetic.`;

  console.log(`🎨 Generating image for ${agent.name} (${agent.role})...`);

  try {
    const response = await client.images.generate({
      model: deploymentName || "dall-e-3",
      prompt,
      n: 1,
      size: "1024x1024",
      quality: "standard",
      style: "vivid",
    });

    const imageUrl = response.data[0]?.url;
    if (!imageUrl) {
      console.error(`  ❌ No image URL returned for ${agent.id}`);
      return null;
    }

    const imageResponse = await fetch(imageUrl);
    const arrayBuffer = await imageResponse.arrayBuffer();
    const buffer = Buffer.from(arrayBuffer);

    const outputPath = path.join(OUTPUT_DIR, `${agent.id}.png`);
    fs.writeFileSync(outputPath, buffer);
    console.log(`  ✅ Saved: ${outputPath}`);

    return outputPath;
  } catch (error) {
    console.error(`  ❌ Error generating ${agent.id}:`, error);
    return null;
  }
}

async function main() {
  const azureEndpoint = process.env.AZURE_OPENAI_ENDPOINT;
  const azureApiKey = process.env.AZURE_OPENAI_API_KEY;
  const azureDeployment =
    process.env.AZURE_OPENAI_DALLE_DEPLOYMENT || "dall-e-3";
  const openaiApiKey = process.env.OPENAI_API_KEY;

  let client: OpenAI | AzureOpenAI;
  let deploymentName: string | undefined;

  if (azureEndpoint && azureApiKey) {
    console.log("🔷 Using Azure OpenAI...");
    // Extract base endpoint if full URL was provided
    const baseEndpoint = azureEndpoint.replace(/\/openai\/.*$/, '');
    console.log(`   Endpoint: ${baseEndpoint}`);
    console.log(`   Deployment: ${azureDeployment}`);
    client = new AzureOpenAI({
      endpoint: baseEndpoint,
      apiKey: azureApiKey,
      apiVersion: "2024-02-01",
    });
    deploymentName = azureDeployment;
  } else if (openaiApiKey) {
    console.log("🟢 Using direct OpenAI...");
    client = new OpenAI({ apiKey: openaiApiKey });
  } else {
    console.error("❌ No API credentials found!");
    console.log("\nSet environment variables:");
    console.log(
      "  AZURE_OPENAI_ENDPOINT + AZURE_OPENAI_API_KEY (for Azure OpenAI)"
    );
    console.log("  OPENAI_API_KEY (for direct OpenAI)");
    process.exit(1);
  }

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
    `\n🤖 Generating ${agentsToGenerate.length} robot agent image(s)...\n`
  );

  const results = { success: [] as string[], failed: [] as string[] };

  for (const agent of agentsToGenerate) {
    const result = await generateImage(client, agent, deploymentName);
    if (result) {
      results.success.push(agent.id);
    } else {
      results.failed.push(agent.id);
    }
    await new Promise((resolve) => setTimeout(resolve, 2000));
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
