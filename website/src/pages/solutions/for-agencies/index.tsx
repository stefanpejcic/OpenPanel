import React from "react";
import { SolutionAccent, SolutionPage } from "@site/src/refine-theme/solutions-page";

export default function ForAgencies() {
    return (
        <SolutionPage
            slug="for-agencies"
            eventPrefix="solutions_agencies"
            metaTitle="OpenPanel for Agencies: Secure Client Hosting, No Add-ons"
            metaDescription="Host every client on your own server: isolation, WAF, malware scanning, backups and a WordPress manager in one €14.95/month licence."
            eyebrow="OpenPanel for agencies"
            heroTitle={
                <>
                    Host every client.{" "}
                    <SolutionAccent>Win back the billable hours.</SolutionAccent>
                </>
            }
            heroText={
                <>
                    Hacked plugins, one client's traffic spike taking down
                    the rest, restores at midnight. OpenPanel gives every
                    client their own isolated space, a firewall, backups and a
                    WordPress manager, all in one licence, so your team gets
                    back to work clients actually pay for.
                </>
            }
            stats={[
                { value: "€14.95", label: "per server per month, any number of clients" },
                { value: "1", label: "isolated environment per client" },
                { value: "0", label: "extra licences for security or backups" },
                { value: "1 panel", label: "for every WordPress site you manage" },
            ]}
            groups={[
                {
                    title: "Built for the way agencies work",
                    subtitle:
                        "Agencies don't want to be sysadmins. These are the server problems that eat your week, handled inside the panel.",
                    items: [
                        {
                            title: "One client's traffic spike takes everyone down",
                            problem:
                                "A campaign launch or a bad plugin on one site maxes out the server, and suddenly every client is calling.",
                            solvedBy: ["User contexts", "Service containers", "Resource limits"],
                            product: "An isolated environment per client",
                            description:
                                "Give each client their own account in its own isolated context, with every service (web server, PHP, database) in a separate container and its own CPU, memory and disk limits. Their sites, databases and files are separated from every other client.",
                            instead: "account isolation",
                            outcome:
                                "One client's problem stays one client's problem",
                            link: "/docs/articles/containers/how-does-openpanel-isolate-users/",
                            image: "/img/openpanel-screenshots/advanced/resource_usage-page",
                            imageAlt: "Resource usage page with CPU and RAM of the account",
                        },
                        {
                            title: "Updating dozens of WordPress sites by hand",
                            problem:
                                "Logging in to every wp-admin to update plugins, check versions and toggle debug mode is hours of unbillable work every month.",
                            solvedBy: ["WordPress Manager"],
                            product: "WordPress manager",
                            description:
                                "Install, update and manage WordPress sites without opening wp-admin. Update plugins and themes, toggle debugging, define plugin and theme sets for new sites, and import existing installs.",
                            instead: "a WordPress toolkit",
                            outcome: "Maintenance rounds in minutes, not afternoons",
                            link: "/docs/panel/applications/wordpress/",
                            image: "/img/openpanel-screenshots/applications/sites-list",
                            imageAlt: "Site manager listing WordPress, Node.js and Python sites",
                        },
                        {
                            title: "Changes break production",
                            problem:
                                "A plugin update or a redesign goes straight to the live site, and you find out from the client.",
                            solvedBy: ["Clone", "Backups"],
                            product: "Staging copies and one-click restore",
                            description:
                                "Clone a WordPress site to a staging location, test there, and keep scheduled backups of files and databases you can restore from the panel when something goes wrong.",
                            instead: "a staging or backup plugin",
                            outcome: "Ship changes without holding your breath",
                            link: "/docs/articles/websites/wordpress-staging-site/",
                            image: "/img/openpanel-screenshots/applications/wordpress-clone",
                            imageAlt: "Clone a WordPress site to a staging location",
                        },
                        {
                            title: "Old client sites you can't safely touch",
                            problem:
                                "Legacy sites need an old PHP version while new builds need the latest, and one server has to run both.",
                            solvedBy: ["PHP per domain", "Node.js", "Python"],
                            product: "The right stack for every site",
                            description:
                                "Pick the PHP version per domain, edit php.ini and extensions, and run Node.js, Python or Ruby apps next to classic PHP sites, with Git deployments for your own builds.",
                            instead: "a separate dev server",
                            outcome: "Legacy and modern sites, side by side",
                            link: "/docs/panel/php/domains/",
                            image: "/img/openpanel-screenshots/php/domains-list",
                            imageAlt: "PHP version per domain",
                        },
                    ],
                },
                {
                    title: "The security layer, included",
                    subtitle:
                        "Every hacked client site costs you a day of cleanup and a little bit of trust. OpenPanel blocks, scans and recovers without extra tools.",
                    items: [
                        {
                            title: "Hacked sites and the cleanup that follows",
                            problem:
                                "Vulnerable plugins are the number one way client sites get compromised, and cleanup is never billable.",
                            solvedBy: ["Coraza WAF", "OWASP CRS"],
                            product: "Web application firewall",
                            description:
                                "CorazaWAF with the OWASP Core Rule Set blocks SQL injection, XSS and other common attacks before they reach WordPress. The WordPress app profile prevents false blocks while editing.",
                            instead: "a WAF or security plugin",
                            outcome: "Most attacks never reach the site",
                            link: "/docs/panel/security/waf/",
                            image: "/img/openpanel-screenshots/advanced/waf_domain-profiles",
                            imageAlt: "Web firewall app profiles for WordPress, Drupal, Nextcloud and more",
                        },
                        {
                            title: "Malware hiding in uploads and themes",
                            problem:
                                "A backdoor in a nulled theme or an upload folder can sit unnoticed until Google flags the site.",
                            solvedBy: ["ClamAV", "ImunifyAV"],
                            product: "Malware scanning with quarantine",
                            description:
                                "Scheduled scans find infected files and move them to quarantine automatically, and you can scan any folder on demand from the panel.",
                            instead: "a malware scanner",
                            outcome: "Infections caught before clients see them",
                            link: "/docs/panel/security/malware-scanner/",
                            image: "/img/openpanel-screenshots/files/malware_logs-list",
                            imageAlt: "Malware scan logs with clean and infected results",
                        },
                        {
                            title: "Clients need access, but not to everything",
                            problem:
                                "Sharing one login across clients and freelancers is a security incident waiting to happen.",
                            solvedBy: ["2FA", "Passkeys", "Branding"],
                            product: "Separate, secure logins under your brand",
                            description:
                                "Every client gets their own login, protected with 2FA or passkeys, with an activity log and active sessions you can review. The panel shows your agency's name and logo.",
                            instead: "white-label branding",
                            outcome: "Clients self-serve, you stay in control",
                            link: "/docs/admin/settings/openpanel/",
                            image: "/img/openadmin-screenshots/settings/openpanel-branding",
                            imageAlt: "Branding settings with brand name, logo, favicon and logout URL",
                        },
                        {
                            title: "A full server outage or a bad deploy",
                            problem:
                                "When something big goes wrong, you need every client back online, not a restore ticket per site.",
                            solvedBy: ["Account backups", "Remote storage"],
                            product: "Automatic off-server backups",
                            description:
                                "Back up all client accounts on one schedule to S3-compatible storage, SFTP or another remote destination, and restore whole accounts when you need to.",
                            instead: "a backup service",
                            outcome: "A recovery plan that doesn't depend on luck",
                            link: "/docs/admin/backups/user/",
                            image: "/img/openpanel-screenshots/files/backups-page",
                            imageAlt: "Backups page for connecting a remote destination",
                        },
                    ],
                },
            ]}
            columnsTitle="Spend less time on servers"
            columns={[
                {
                    title: "Less firefighting, more billable time",
                    text: "Isolation, WAF and malware scanning work in the background, so fewer late-night emergencies pull your developers off client work.",
                },
                {
                    title: "Protect your reputation",
                    text: "Your clients judge you by their site being fast and online. A secure, isolated setup keeps one bad plugin from becoming your problem with every client.",
                },
                {
                    title: "Many clients, one affordable server",
                    text: "Host as many client accounts as the hardware allows for one flat price, and turn hosting into recurring revenue instead of a cost.",
                },
            ]}
            compareTitle="Everything an agency server needs"
            compareSubtitle="What you would normally piece together from separate plugins, services and licences."
            compareRows={[
                { feature: "Per-client isolation", typical: "Separate licence" },
                { feature: "Web application firewall", typical: "Plugin or licence" },
                { feature: "Malware scanning", typical: "Plugin or licence" },
                { feature: "Off-server backups", typical: "Separate service" },
                { feature: "WordPress manager", typical: "Paid add-on" },
                { feature: "Staging and cloning", typical: "Paid add-on" },
                { feature: "Your branding", typical: "Paid add-on" },
            ]}
            faqTitle="Questions agencies ask first"
            faq={[
                {
                    question: "Do I still need security plugins on every WordPress site?",
                    answer: "The Coraza WAF and malware scanner protect every site at the server level, so you don't depend on each client keeping a security plugin installed and updated. You should still keep WordPress, themes and plugins updated.",
                },
                {
                    question: "Does OpenPanel replace WordPress updates?",
                    answer: "No, but it makes them faster: the WordPress manager lets you update plugins and themes across sites from one place without logging in to each wp-admin.",
                },
                {
                    question: "Can I give clients their own login?",
                    answer: "Yes. Each client gets their own OpenPanel account and login, isolated from other clients, with 2FA and passkeys. You decide which features they can use and the panel carries your branding.",
                },
                {
                    question: "Can I host legacy sites on old PHP versions?",
                    answer: "Yes. The PHP version is set per domain, so an old client site and a brand new build can run side by side on the same server.",
                },
                {
                    question: "How many client sites can I host on one licence?",
                    answer: "As many as your server can handle. The Enterprise licence is €14.95 per server per month with unlimited accounts, domains and websites.",
                },
            ]}
            ctaTitle="Spend less time firefighting. More time on clients."
            ctaText="Try OpenPanel Enterprise free for 30 days and move your client sites to a server that protects and backs them up on its own."
        />
    );
}
