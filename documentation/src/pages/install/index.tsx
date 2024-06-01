import React, { useState, useEffect } from "react";
import Head from "@docusaurus/Head";
import { BlogFooter } from "@site/src/refine-theme/blog-footer";
import { CommonHeader } from "@site/src/refine-theme/common-header";
import { CommonLayout } from "@site/src/refine-theme/common-layout";
import clsx from "clsx";

const Install: React.FC = () => {
    const [installOptions, setInstallOptions] = useState({
        hostname: { value: "", description: "Set the hostname." },
        version: { value: "", description: "Set a custom OpenPanel version to be installed." },
        skip_requirements: { value: false, description: "Skip the requirements check." },
        skip_panel_check: { value: false, description: "Skip checking if existing panels are installed." },
        skip_apt_update: { value: false, description: "Skip the APT update." },
        overlay2: { value: false, description: "Enable overlay2 storage driver instead of device-mapper." },
        skip_firewall: { value: false, description: "Skip UFW setup UFW - Only do this if you will set another Firewall manually!" },
        skip_images: { value: false, description: "Skip installing openpanel/nginx and openpanel/apache docker images." },
        skip_blacklists: { value: false, description: "Do not set up IP sets and blacklists." },
        skip_ssl: { value: false, description: "Skip SSL setup." },
        with_modsec: { value: false, description: "Enable ModSecurity for Nginx." },
        ips: { value: "", description: "Whitelist IP addresses of OpenPanel Support Team." },
        no_ssh: { value: false, description: "Disable port 22 and whitelist the IP address of user installing the panel." },
        enable_ftp: { value: false, description: "Install FTP (experimental)." },
        enable_mail: { value: false, description: "Install Mail (experimental)." },
        post_install: { value: "", description: "Specify the post install script path." },
        screenshots: { value: "local", description: "Set the screenshots API URL." },
        debug: { value: false, description: "Display debug information during installation." },
        repair: { value: false, description: "Retry and overwrite everything." },
    });

    const [latestVersion, setLatestVersion] = useState<string>("");

    useEffect(() => {
        fetch("https://update.openpanel.co/")
            .then(response => response.text())
            .then(data => {
                // Extract the latest version from the plain text
                setLatestVersion(data.trim());
                // Set default version to the latest version
                setInstallOptions(prevState => ({
                    ...prevState,
                    version: { ...prevState.version, value: data.trim() },
                }));
            })
            .catch(error => console.error("Error fetching latest version:", error));
    }, []);

    const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
        const { name, value, type, checked } = e.target;
        setInstallOptions(prevState => ({
            ...prevState,
            [name]: { ...prevState[name], value: type === "checkbox" ? checked : value },
        }));
    };

    const generateInstallCommand = () => {
        let command = "bash <(curl -sSL https://get.openpanel.co/)";
        for (const option in installOptions) {
            if (installOptions[option].value) {
                if (typeof installOptions[option].value === "boolean") {
                    command += ` --${option}`;
                } else {
                    command += ` --${option}=${installOptions[option].value}`;
                }
            }
        }
        return command;
    };

    return (
        <CommonLayout>
            <Head title="Install | OpenPanel">
                <html data-page="install" data-customized="true" />
            </Head>
            <div className="refine-prose">
                <CommonHeader hasSticky={true} />

                <div className="flex-1 flex flex-col pt-8 lg:pt-16 pb-32 max-w-[800px] w-full mx-auto px-2">
                    <h1>Installation</h1>

                    <div className="mb-8">
                        <h2>Install Command:</h2>
                        <pre>{generateInstallCommand()}</pre>
                    </div>

                    <div className="mb-8">
                        <h2>Advanced Install Settings:</h2>
                        <div className="grid grid-cols-1 md:grid-cols-2 gap-4">
                            {/* Input fields for advanced install settings */}
                            {Object.entries(installOptions).map(([key, value]) => (
                                <div key={key}>
                                    <label htmlFor={key}>{key.replace(/_/g, ' ').toUpperCase()}:</label>
                                    <span>{value.description}</span>
                                    {key === "version" ? (
                                        <input
                                            type="text"
                                            id={key}
                                            name={key}
                                            value={latestVersion}
                                            readOnly
                                        />
                                    ) : (
                                        typeof value.value === "boolean" ? (
                                            <input
                                                type="checkbox"
                                                id={key}
                                                name={key}
                                                checked={value.value}
                                                onChange={handleInputChange}
                                            />
                                        ) : (
                                            <input
                                                type="text"
                                                id={key}
                                                name={key}
                                                value={value.value}
                                                onChange={handleInputChange}
                                            />
                                        )
                                    )}
                                </div>
                            ))}
                        </div>
                    </div>
                </div>
                <BlogFooter />
            </div>
        </CommonLayout>
    );
};

export default Install;
