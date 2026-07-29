import os
import pathlib
from google.adk.agents import Agent
from google.adk.skills import models, SkillToolset

# 1. Define Meta-Skill (The Skill Factory)
skill_creator = models.Skill(
    frontmatter=models.Frontmatter(
        name="skill-creator",
        description="Creates new ADK-compatible skill definitions following agentskills.io spec.",
    ),
    instructions=(
        "When asked to create a new skill, generate a complete SKILL.md file.\n"
        "Read `references/agentskills_spec.md` for format requirements.\n"
        "Rules:\n"
        "1. Name must be kebab-case, max 64 chars\n"
        "2. Description under 1024 chars\n"
        "3. Clear, step-by-step instructions\n"
        "4. Place deep domain details in references/\n"
    ),
    resources=models.Resources(
        references={
            "agentskills_spec.md": "# Agent Skills Specification (agentskills.io)..."
        }
    ),
)

# 2. Assemble SkillToolset and Agent
skill_toolset = SkillToolset(skills=[skill_creator])

root_agent = Agent(
    model="gemini-2.5-flash",
    name="self_extending_agent",
    description="Agent capable of loading and writing its own skills on demand.",
    tools=[skill_toolset],
)
