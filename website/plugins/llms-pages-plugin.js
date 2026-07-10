const fs = require("fs");
const path = require("path");

// Some pages don't follow the CommonLayout description / Head title convention
// (the homepage uses its own <title> tag, the integrations page has no page-level
// meta at all), so their title/description is supplied here instead of parsed.
const OVERRIDES = {
    "": {
        title: "OpenPanel",
        description:
            "A highly customizable web hosting control panel built around Docker containers.",
    },
    features: {
        title: "Integrations & Features",
        description:
            "Browse OpenPanel's integrations and features - billing, DNS, mail, backups, and more.",
    },
    docs: {
        title: "Documentation",
        description:
            "OpenPanel and OpenAdmin documentation - installation, opencli commands, and server management.",
    },
};

// Mirrors the site's own footer navigation (src/refine-theme/footer-data.tsx),
// grouped the way Cloudflare structures llms.txt: a blurb per section, then links.
const SECTIONS = [
    {
        title: "Product",
        blurb: "The core control panel: a Docker-isolated environment per hosting user, managed through OpenPanel (user-facing) and OpenAdmin (server administration).",
        slugs: ["", "features", "install", "demo"],
    },
    {
        title: "Editions & Pricing",
        blurb: "OpenPanel ships as a free Community edition and a paid Enterprise edition, priced per server rather than per hosting account. Billing-software integrations automate account provisioning.",
        slugs: [
            "community",
            "enterprise",
            "licenses-for-partners",
            "billing",
            "blesta-module",
            "fossbilling-module",
            "whmcs-module",
            "refund-policy",
        ],
    },
    {
        title: "Compare OpenPanel",
        blurb: "How OpenPanel's Docker-per-user model differs from CloudLinux-based and traditional shared-hosting control panels.",
        slugs: [
            "cpanel-alternative",
            "plesk-alternative",
            "directadmin-alternative",
            "cyberpanel-alternative",
        ],
    },
    {
        title: "Company",
        blurb: "Background on the project, its roadmap, live usage data, and hosting providers running OpenPanel for their customers.",
        slugs: ["about", "roadmap", "statistics", "assets", "hosting-providers"],
    },
    {
        title: "Support & Legal",
        blurb: "Where to get help, report bugs, or apply a patch, plus the license terms OpenPanel is distributed under.",
        slugs: ["support", "docs", "patches", "privacy-policy", "LICENSE"],
    },
];

const EXTERNAL_RESOURCES = [
    {
        title: "Blog",
        url: "https://openpanel.com/blog",
        description: "Release notes, feature announcements, and hosting how-tos.",
    },
    {
        title: "GitHub",
        url: "https://github.com/stefanpejcic/openpanel",
        description:
            "Source repository, issue tracker, and CONTRIBUTING guide. OpenPanel UI/OpenAdmin are distributed under a EULA; opencli and configuration files are CC BY-NC.",
    },
    {
        title: "Discord",
        url: "https://discord.com/invite/7bNY8fANqF",
        description: "Community chat for support and discussion.",
    },
    {
        title: "Reddit",
        url: "https://www.reddit.com/r/openpanelco/",
        description: "Community discussion board.",
    },
];

module.exports = function llmsPagesPlugin(context, options = {}) {
    const pagesDir = path.join(context.siteDir, "src", "pages");
    const staticDir = path.join(context.siteDir, "static");
    const outputFile = path.join(staticDir, options.outputFile || "llms.txt");
    const siteUrl = (context.siteConfig.url || "").replace(/\/$/, "");

    function extractMeta(content) {
        const titleMatch =
            content.match(/<Head title="([^"]+)"/) ||
            content.match(/const title = "([^"]+)"/);
        const descriptionMatch = content.match(
            /<CommonLayout\s+description="([^"]+)"/,
        );
        return {
            title: titleMatch ? titleMatch[1] : null,
            description: descriptionMatch ? descriptionMatch[1] : null,
        };
    }

    function findIndexFile(dir) {
        return ["index.tsx", "index.jsx"]
            .map((name) => path.join(dir, name))
            .find((file) => fs.existsSync(file));
    }

    function toUrl(slug) {
        return slug ? `${siteUrl}/${slug}/` : `${siteUrl}/`;
    }

    function getPage(slug) {
        const indexFile = findIndexFile(path.join(pagesDir, slug));
        if (!indexFile) return null;

        const meta = extractMeta(fs.readFileSync(indexFile, "utf8"));
        const override = OVERRIDES[slug];
        const title = (override && override.title) || meta.title;
        const description = (override && override.description) || meta.description;

        if (!title) return null;

        return {
            title: title.split("|")[0].trim(),
            description,
            url: toUrl(slug),
        };
    }

    function formatLink({ title, description, url }) {
        return description
            ? `- [${title}](${url}): ${description}`
            : `- [${title}](${url})`;
    }

    function generateContent() {
        const lines = [
            "# OpenPanel",
            "",
            "> OpenPanel is a web hosting control panel built on Docker. Every hosting user gets a fully isolated environment - dedicated web server, database, PHP-FPM, private network, and storage - delivering VPS-grade isolation on shared hardware. It ships as a free Community edition and a paid Enterprise edition for hosting providers, priced per server rather than per account.",
            "",
            "This is the main OpenPanel website. For technical documentation, see openpanel.com/docs.",
            "",
            "## About OpenPanel",
            "",
            "OpenPanel started in 2023 as a LAMP-stack panel and moved to one Docker container per user service (web, database, mail), giving each hosting user real isolation without the overhead of a container per website.",
            "",
            "Key facts:",
            "- Free Community edition, plus Enterprise from a fixed €14.95/month per server - no per-account fees",
            "- One Docker container per user service: web server, database, PHP-FPM, mail",
            "- Supports Ubuntu, Debian, AlmaLinux, RockyLinux, and CentOS on AMD64 and ARM",
            "- REST API with 160+ endpoints, plus a native MCP server for AI agent management (since v1.7.65)",
            "- Billing automation for WHMCS, FOSSBilling, and Blesta",
            "- Not open source: OpenPanel UI and OpenAdmin are distributed under a EULA; opencli and configuration files are CC BY-NC (non-commercial)",
        ];

        SECTIONS.forEach((section) => {
            const items = section.slugs.map(getPage).filter(Boolean);
            if (items.length === 0) return;

            lines.push("", "---", "", `## ${section.title}`, "");
            if (section.blurb) lines.push(section.blurb, "");
            items.forEach((item) => lines.push(formatLink(item)));
        });

        lines.push(
            "",
            "---",
            "",
            "## External Resources",
            "",
            ...EXTERNAL_RESOURCES.map(formatLink),
            "",
            "## Optional",
            "",
            `- [Full Context Version](${siteUrl}/llms-full.txt): Complete OpenPanel and OpenAdmin documentation inline, for larger context windows`,
        );

        if (!fs.existsSync(staticDir)) {
            fs.mkdirSync(staticDir, { recursive: true });
        }
        fs.writeFileSync(outputFile, `${lines.join("\n")}\n`);
        console.log(`Generated: ${outputFile}`);
    }

    return {
        name: "llms-pages-plugin",

        async loadContent() {
            generateContent();
        },

        extendCli(cli) {
            cli.command("generate-llms-pages-txt")
                .description("Generate the pages-only llms.txt file")
                .action(() => generateContent());
        },
    };
};
