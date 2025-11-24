import React, { useState, ChangeEvent } from "react";
import Head from "@docusaurus/Head";
import { BlogFooter } from "@site/src/refine-theme/blog-footer";
import { CommonHeader } from "@site/src/refine-theme/common-header";
import { CommonLayout } from "@site/src/refine-theme/common-layout";

interface InstallOption {
    value: string | boolean;
    description: string;
}

type InstallOptions = Record<string, InstallOption>;

const defaultOptions: InstallOptions = {
    key: { value: "", description: "Set Enterprise license key." },
    domain: { value: "", description: "Set the domain to be used for accessing panels." },
    email: { value: "", description: "Email address to receive admin logins and future notifications." },
    username: { value: "", description: "Set admin username (by default random generated)." },
    password: { value: "", description: "Set admin password (by default random generated)." },
    "skip-firewall": { value: false, description: "Don't setup Sentinel Firewall (CSF)" },
    "skip-dns-server": { value: false, description: "Don't setup local BIND9 DNS server" },
    "imunifyav": { value: false, description: "Install and Setup ImunifyAV." },
    "no-waf": { value: false, description: "Do not install CorazaWAF and disable it for new domains." },
    "post-install": { value: "", description: "Specify the post install script path or URL." },
    swap: { value: "", description: "Set size in GB for the swap partition." },
    screenshots: { value: "", description: "Set the screenshots API URL." },
    "skip-requirements": { value: false, description: "Skip the requirements check." },
    "skip-panel-check": { value: false, description: "Skip checking if existing panels are installed." },
    "skip-apt-update": { value: false, description: "Skip the APT update." },
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
            if (typeof config.value === "boolean") {
                if (config.value) {
                    command += ` --${option}`;
                }
            } else if (config.value.trim() !== "") {
                if (option === "username" || option === "password" || option === "post-install" || option === "screenshots") {
                    command += ` --${option}='${config.value}'`;
                } else {
                    command += ` --${option}=${config.value}`;
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
                                    <input
                                        type={
                                            key === "email" ? "email" :
                                            typeof config.value === "boolean" ? "checkbox" : "text"
                                        }
                                        id={key}
                                        name={key}
                                        {...(typeof config.value === "boolean"
                                            ? { checked: config.value }
                                            : { value: config.value })}
                                        onChange={handleInputChange}
                                        pattern={
                                            key === "username" || key === "password"
                                                ? "^[a-zA-Z0-9]+$"
                                                : key === "domain"
                                                ? "^(?!-)(?:[A-Za-z0-9-]{1,63}\\.)+[A-Za-z]{2,}$"
                                                : key === "key"
                                                ? "^enterprise-.*$"
                                                : undefined
                                        }
                                        title={
                                            key === "username" || key === "password"
                                                ? "Only letters and numbers are allowed"
                                                : key === "domain"
                                                ? "Enter a valid domain, e.g. openpanel.server.com"
                                                : key === "key"
                                                ? "License key must start with 'enterprise-'"
                                                : undefined
                                        }
                                        className="mr-2 w-64 px-1 py-2 text-sm dark:placeholder-gray-500 placeholder-gray-400 dark:text-gray-500 text-gray-400 dark:bg-gray-900 bg-white border dark:border-gray-700 border-gray-300 rounded-lg resize-none"
                                    />
                                    <div>
                                        <label htmlFor={key} className="font-bold">
                                            {key.replace(/-/g, " ").toUpperCase()}
                                        </label>
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
