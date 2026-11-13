import React from "react";
import {
    SolutionAccent,
    SolutionBillingStrip,
    SolutionPage,
} from "@site/src/refine-theme/solutions-page";

export default function ForHostingProviders() {
    return (
        <SolutionPage
            slug="for-hosting-providers"
            eventPrefix="solutions_hosting"
            metaTitle="OpenPanel for Hosting Providers: One Licence, Everything Built In"
            metaDescription="OpenPanel gives hosting providers per-user isolation, a web application firewall, malware scanning and backups in one €14.95/month licence - no add-on licences, no per-account fees."
            eyebrow="OpenPanel for hosting providers"
            heroTitle={
                <>
                    Fewer licences.{" "}
                    <SolutionAccent>More margin per server.</SolutionAccent>
                </>
            }
            heroText={
                <>
                    Most panels are just the start of the bill: isolation,
                    WAF, backups and security each come as their own licence,
                    usually priced per account. OpenPanel ships all of it in
                    one flat licence, so every new customer on a server is
                    pure margin.
                </>
            }
            stats={[
                { value: "€14.95", label: "per server per month, flat" },
                { value: "€0", label: "per-account or per-site fees" },
                { value: "1", label: "licence instead of an add-on stack" },
                { value: "30 days", label: "free trial, no credit card" },
            ]}
            groups={[
                {
                    title: "The foundation, already inside",
                    subtitle:
                        "The problems that cost hosts the most tickets and churn, solved by features that are part of OpenPanel, not bolted on.",
                    items: [
                        {
                            title: "One noisy account slows the whole server",
                            problem:
                                "A traffic spike or a runaway script on one site eats CPU and RAM, and every other customer on the node opens a ticket.",
                            solvedBy: ["User contexts", "Service containers", "Hosting plans"],
                            product: "Isolated users, containerised services",
                            description:
                                "Every account runs in its own isolated context, so users can't see or reach each other. Inside it, each service (web server, PHP, database, Redis) runs in its own container on segmented internal networks. CPU, memory, disk and inode limits are set per hosting plan, so one account can never take the server down for the rest.",
                            instead: "kernel-level account isolation",
                            outcome:
                                "Bad neighbours stay contained, other customers never notice",
                            link: "/docs/articles/containers/how-does-openpanel-isolate-users/",
                            image: "/img/openadmin-screenshots/plans/hosting_plans-edit",
                            imageAlt: "Hosting plan editor with disk, inode, CPU, memory and port speed limits",
                        },
                        {
                            title: "Sites get hacked through vulnerable plugins",
                            problem:
                                "SQL injection and XSS attempts hit every WordPress and Joomla site daily, and each successful one becomes a cleanup ticket.",
                            solvedBy: ["Coraza WAF", "OWASP CRS"],
                            product: "Web application firewall",
                            description:
                                "CorazaWAF with the OWASP Core Rule Set inspects every request before it reaches a site. Protect or monitor-only mode per domain, app profiles for popular CMSs to avoid false blocks, and per-domain logs of what was stopped.",
                            instead: "a WAF or security suite",
                            outcome:
                                "Attacks are blocked at the edge, rules update automatically",
                            link: "/docs/admin/security/waf/",
                            image: "/img/openadmin-screenshots/security/waf-page",
                            imageAlt: "Coraza web firewall overview in OpenAdmin",
                        },
                        {
                            title: "Malware spreads before anyone notices",
                            problem:
                                "Infected files sit on the server for weeks, then the IP gets blacklisted and the abuse reports start.",
                            solvedBy: ["ClamAV", "ImunifyAV"],
                            product: "Malware scanning and quarantine",
                            description:
                                "Scheduled ClamAV scans move infected files to quarantine automatically, users can run their own scans, and ImunifyAV is included and pre-wired into OpenAdmin.",
                            instead: "a malware scanner",
                            outcome:
                                "Infections are found and isolated without manual digging",
                            link: "/docs/panel/security/malware-scanner/",
                            image: "/img/openpanel-screenshots/files/quarantine-list",
                            imageAlt: "Malware quarantine with infected files and restore, mark as safe and delete actions",
                        },
                        {
                            title: "Backups are a separate product to buy and babysit",
                            problem:
                                "Customers expect restores, but the backup tool is another licence, another dashboard and another thing to keep running.",
                            solvedBy: ["Account backups", "System backups"],
                            product: "Built-in account and server backups",
                            description:
                                "Back up every account centrally on your schedule, or let customers manage their own backups and restores. Send them to S3-compatible storage, SFTP and other remote destinations, and back up the panel configuration itself for disaster recovery.",
                            instead: "a backup add-on",
                            outcome:
                                "Restores are self-service, fewer \"please restore my site\" tickets",
                            link: "/docs/admin/backups/user/",
                            image: "/img/openpanel-screenshots/files/backups_destination-page",
                            imageAlt: "Backup destinations: S3-compatible, WebDAV, SSH, Azure and Dropbox",
                        },
                    ],
                },
                {
                    title: "Built for running a hosting business",
                    subtitle:
                        "Security at the server level, plus the tools to sell, resell and migrate, without another vendor in the chain.",
                    items: [
                        {
                            title: "Brute force and port scans hit the server around the clock",
                            problem:
                                "Without a properly configured firewall, every server is a target for SSH, mail and panel login attacks.",
                            solvedBy: ["Sentinel Firewall", "IP blocker"],
                            product: "Server firewall",
                            description:
                                "Sentinel Firewall, a maintained fork of CSF, is installed and managed from OpenAdmin. Admins and users can block IPs, and logins support 2FA and passkeys.",
                            instead: "a firewall add-on",
                            outcome: "Hardened from the first boot",
                            link: "/docs/admin/security/firewall/",
                            image: "/img/openpanel-screenshots/advanced/ip-blocker-form",
                            imageAlt: "IP blocker with blocked addresses and ranges",
                        },
                        {
                            title: "Slow sites drive customers to the competition",
                            problem:
                                "Page speed is what customers compare, and a slow shared server is the first thing they blame.",
                            solvedBy: ["Varnish", "Redis", "Memcached"],
                            product: "Per-user web stack and caching",
                            description:
                                "Each user picks Nginx, Apache or OpenLiteSpeed, puts Varnish in front, and adds Redis, Valkey or Memcached for object caching. Multiple PHP versions per domain, plus Node.js, Python and more.",
                            instead: "a caching or acceleration add-on",
                            outcome: "Fast sites without a premium tier to sell",
                            link: "/docs/panel/caching/caching/",
                            image: "/img/openpanel-screenshots/caching/redis-page",
                            imageAlt: "Redis service page with status and connection details",
                        },
                        {
                            title: "Resellers need their own panel and your brand",
                            problem:
                                "Reseller hosting usually means a higher licence tier, and white-labelling costs extra on top.",
                            solvedBy: ["Resellers", "Branding"],
                            product: "Resellers and white-label",
                            description:
                                "Create resellers with their own accounts, packages and disk limits. Set your own brand name, logo and logout URL so customers only ever see your company.",
                            instead: "a reseller or white-label tier",
                            outcome: "A new product line from the same servers",
                            link: "/docs/admin/accounts/resellers/",
                            image: "/img/openadmin-screenshots/accounts/resellers-list",
                            imageAlt: "Resellers list with accounts, storage and hosting plans",
                        },
                        {
                            title: "Moving customers in is slow and risky",
                            problem:
                                "Migrating from another panel, or between your own servers, is where weekends disappear.",
                            solvedBy: ["cPanel import", "Account transfer"],
                            product: "Migrations and multi-server",
                            description:
                                "Import accounts straight from cPanel backups, transfer accounts between OpenPanel servers, and manage several servers as a cluster with DNS clustering.",
                            instead: "a migration tool",
                            outcome: "Switch panels without losing a customer",
                            link: "/docs/admin/backups/cpanel/",
                            image: "/img/openadmin-screenshots/advanced/cpanel-form",
                            imageAlt: "Import account from an OpenPanel, cPanel or CyberPanel backup",
                        },
                    ],
                },
            ]}
            columnsTitle="Why the numbers work"
            columns={[
                {
                    title: "Flat price, no account tiers",
                    text: "€14.95 per server whether it hosts 10 accounts or 1,000. You never get pushed to a higher tier because you grew.",
                },
                {
                    title: "One vendor, one update cycle",
                    text: "Isolation, WAF, backups and the panel are released and tested together, so updates don't break each other and support isn't bounced between vendors.",
                },
                {
                    title: "Runs on cheaper hardware",
                    text: "Native ARM (AArch64) support and lightweight containers let you use cost-efficient servers and pack more customers per node.",
                },
            ]}
            compareTitle="What's in the box"
            compareSubtitle="The layers a production hosting server needs, and how they're usually licensed."
            compareRows={[
                { feature: "Per-account isolation", typical: "Separate licence" },
                { feature: "Web application firewall", typical: "Separate licence" },
                { feature: "Malware scanning", typical: "Separate licence" },
                { feature: "Account backups", typical: "Separate licence" },
                { feature: "WordPress manager", typical: "Paid add-on" },
                { feature: "App auto-installer", typical: "Paid add-on" },
                { feature: "Reseller accounts", typical: "Higher tier" },
                { feature: "White-label branding", typical: "Paid add-on" },
            ]}
            extra={<SolutionBillingStrip />}
            faqTitle="Questions hosting providers ask first"
            faq={[
                {
                    question: "Do I need any other licence to run OpenPanel in production?",
                    answer: "No. The Enterprise licence includes isolation, the Coraza WAF, malware scanning, backups, the WordPress manager, resellers and branding. The only optional paid component is LiteSpeed Enterprise, if you choose it over the included Nginx, Apache or OpenLiteSpeed.",
                },
                {
                    question: "Is the price really per server, not per account?",
                    answer: "Yes. €14.95 per server per month with unlimited accounts, domains and websites. There are no per-account tiers to outgrow.",
                },
                {
                    question: "How is isolation different from shared hosting on other panels?",
                    answer: "Every OpenPanel account runs in its own isolated context, so users can't see or reach each other's files and services. Inside that context each service (web server, PHP, MySQL, Redis) runs in its own container, so even a compromised database can't affect the user's other services. Resource limits are set per account, so an overloaded account can't starve others of CPU and memory.",
                },
                {
                    question: "Can I migrate my existing cPanel customers?",
                    answer: "Yes. OpenPanel creates accounts directly from cPanel and CyberPanel backup files, and can transfer accounts between OpenPanel servers.",
                },
                {
                    question: "Does it work with my billing system?",
                    answer: "Free provisioning modules are available for WHMCS, Blesta, FOSSBilling, ClientExec and WiseCP, and everything else can be automated through the REST API or the opencli command line.",
                },
                {
                    question: "Can I try it before buying?",
                    answer: "Yes, Enterprise has a 30-day free trial with no credit card. You can also try the live demo or start with the free Community edition.",
                },
            ]}
            ctaTitle="Turn your licence bill into margin."
            ctaText="Put isolation, WAF and backups on every server for one flat price. Start a free trial or talk to us about moving your fleet."
        />
    );
}
