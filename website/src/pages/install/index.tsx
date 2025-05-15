import React, { useState, useEffect, ChangeEvent } from "react";
import Head from "@docusaurus/Head";
import { BlogFooter } from "@site/src/refine-theme/blog-footer";
import { CommonHeader } from "@site/src/refine-theme/common-header";
import { CommonLayout } from "@site/src/refine-theme/common-layout";
import clsx from "clsx";

interface InstallOption {
    value: string | boolean;
    description: string;
    options?: string[];
}

type InstallOptions = Record<string, InstallOption>;

const defaultOptions: InstallOptions = {
    domain: { value: "", description: "Set the domain to be used for accessing panels." },
    email: { value: "", description: "Email address to receive admin logins and future notifications." },
    username: { value: "", description: "Set admin username (by default random generated)." },
    password: { value: "", description: "Set admin password (by default random generated)." },
    "skip-requirements": { value: false, description: "Skip the requirements check." },
    "skip-panel-check": { value: false, description: "Skip checking if existing panels are installed." },
    "skip-apt-update": { value: false, description: "Skip the APT update." },
    "skip-firewall": { value: false, description: "Don't setup UFW (Only if you will set another Firewall manually)" },
    "skip-dns-server": { value: false, description: "Don't setup DNS server (Only if you will use external NS like Cloudflare)" },
    ufw: { value: false, description: "Install and setup UFW instead of CSF." },
    "no-ssh": { value: false, description: "Disable port 22 and whitelist administrator IP address." },
    "no-waf": { value: false, description: "Do not install CorazaWAF and disable it for new domains." },
    "post-install": { value: "", description: "Specify the post install script path." },
    screenshots: { value: "remote", description: "Set the screenshots API URL.", options: ["local", "remote"] },
    debug: { value: false, description: "Display debug information during installation." },
    repair: { value: false, description: "Retry and overwrite everything." },
};

const Install: React.FC = () => {
    const [installOptions, setInstallOptions] = useState<InstallOptions>(defaultOptions);

    const handleInputChange = (e: ChangeEvent<HTMLInputElement | HTMLSelectElement>) => {
        const { name, value, type, checked } = e.target;
        setInstallOptions(prevState => ({
            ...prevState,
            [name]: { ...prevState[name], value: type === "checkbox" ? checked : value },
        }));
    };

    const generateInstallCommand = () => {
        let command = "bash <(curl -sSL https://openpanel.org)";
        for (const [option, config] of Object.entries(installOptions)) {
            if (option !== "version") {
                if (config.value || ["version", "hostname", "email", "screenshots", "docker-space", "post-install"].includes(option)) {
                    if (option === "screenshots" && config.value === "local") {
                        command += ` --screenshots=local`;
                    } else if (option === "screenshots" && config.value === "remote") {
                        command += ``;
                    } else if (config.value !== "" && config.value !== false) {
                        command += ` --${option}`;
                        if (config.value !== true) {
                            command += `=${config.value}`;
                        }
                    }
                }
            }
        }
        return command;
    };

    return (
        <CommonLayout>
            <Head title="Install command generator | OpenPanel">
                <html data-page="install" data-customized="true" />
            </Head>
            <div className="refine-prose">
                <CommonHeader hasSticky={true} />

                <div className="flex-1 flex flex-col pt-8 lg:pt-16 pb-32 max-w-[800px] w-full mx-auto px-2">
                    <h1>Installation</h1>
                    <p className="mb-0">Connect to your server as root via SSH and run the following command:</p>

                    <div className="mb-0">
                        <pre style={{ whiteSpace: "pre-wrap" }}>{generateInstallCommand()}</pre>
                    </div>
                    
                    <div className="mb-2">
                        <h2>Advanced Install Settings:</h2>
                        <p className="mt-0">Here you can set what shall be installed and configured when installing OpenPanel:</p>
                        <ul>
                            {Object.entries(installOptions).map(([key, config]) => (
                                <li key={key} className="flex items-center mb-2">
                                    {key === "screenshots" ? (
                                        <select
                                            id={key}
                                            name={key}
                                            value={config.value as string}
                                            onChange={handleInputChange}
                                            className="mr-2 w-40 px-1 py-2 text-sm dark:placeholder-gray-500 placeholder-gray-400 dark:text-gray-500 text-gray-400 dark:bg-gray-900 bg-white border dark:border-gray-700 border-gray-300 rounded-lg resize-none"
                                        >
                                            {config.options?.map(option => (
                                                <option key={option} value={option}>{option}</option>
                                            ))}
                                        </select>
                                    ) : (
                                        <input
                                            type={typeof config.value === "boolean" ? "checkbox" : "text"}
                                            id={key}
                                            name={key}
                                            checked={typeof config.value === "boolean" ? config.value : undefined}
                                            value={typeof config.value === "string" ? config.value : undefined}
                                            onChange={handleInputChange}
                                            className="mr-2 w-40"
                                        />
                                    )}
                                    <div>
                                        <label htmlFor={key} className="font-bold">{key.replace(/-/g, ' ').toUpperCase()}</label>
                                        <div>{config.description}</div>
                                    </div>
                                </li>
                            ))}
                        </ul>
                    </div>
                </div>
                <BlogFooter />
            </div>
        </CommonLayout>
    );
};

export default Install;
