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
    version: { value: "", description: "Set a custom OpenPanel version to be installed." },
    email: { value: "", description: "Email address to receive admin logins and future notifications." },
    username: { value: "", description: "Set admin username (by default random generated)." },
    password: { value: "", description: "Set admin password (by default random generated)." },
    "skip-requirements": { value: false, description: "Skip the requirements check." },
    "skip-panel-check": { value: false, description: "Skip checking if existing panels are installed." },
    "skip-apt-update": { value: false, description: "Skip the APT update." },
    "skip-firewall": { value: false, description: "Don't setup UFW (Only if you will set another Firewall manually)" },
    "skip-images": { value: false, description: "Don't install openpanel/nginx and openpanel/apache docker images." },
    "skip-blacklists": { value: false, description: "Do not set up IP sets and blacklists." },
    "skip-ssl": { value: false, description: "Skip SSL setup." },
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
    const [latestVersion, setLatestVersion] = useState<string>("");

    useEffect(() => {
        fetch("https://raw.githubusercontent.com/stefanpejcic/OpenPanel/refs/heads/main/version/latest")
            .then(response => response.text())
            .then(data => {
                const version = data.trim();
                setLatestVersion(version);
                setInstallOptions(prevState => ({
                    ...prevState,
                    version: { ...prevState.version, value: version },
                }));
            })
            .catch(error => console.error("Error fetching latest version:", error));
    }, []);

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
            if (option !== "version" || (option === "version" && config.value !== latestVersion)) {
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
                        <pre className="my-0" style={{ whiteSpace: "pre-wrap" }}>{generateInstallCommand()}</pre>
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
                                            className="mr-2 w-40"
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
