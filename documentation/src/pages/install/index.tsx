import React, { useState, useEffect } from "react";
import Head from "@docusaurus/Head";
import { BlogFooter } from "@site/src/refine-theme/blog-footer";
import { CommonHeader } from "@site/src/refine-theme/common-header";
import { CommonLayout } from "@site/src/refine-theme/common-layout";
import clsx from "clsx";

const Install: React.FC = () => {
    const [installOptions, setInstallOptions] = useState({
        hostname: "",
        version: "",
        skipRequirements: false,
        skipPanelCheck: false,
        skipAptUpdate: false,
        overlay2: false,
        skipFirewall: false,
        skipImages: false,
        skipBlacklists: false,
        skipSSL: false,
        withModsec: false,
        ips: "",
        noSSH: false,
        enableFTP: false,
        enableMail: false,
        postInstall: "",
        screenshots: "local",
        debug: false,
        repair: false,
    });

    const [availableVersions, setAvailableVersions] = useState<string[]>([]);

    useEffect(() => {
        fetch("https://update.openpanel.co/")
            .then(response => response.json())
            .then(data => {
                // Extract the latest version and versions below it
                const latestVersion = data.latest_version;
                const versionsBelow = data.versions_below.slice(0, 5);
                // Combine the latest version and versions below it
                const allVersions = [latestVersion, ...versionsBelow];
                setAvailableVersions(allVersions);
                // Set default version to the latest version
                setInstallOptions(prevState => ({
                    ...prevState,
                    version: latestVersion,
                }));
            })
            .catch(error => console.error("Error fetching available versions:", error));
    }, []);

    const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
        const { name, value, type, checked } = e.target;
        setInstallOptions(prevState => ({
            ...prevState,
            [name]: type === "checkbox" ? checked : value,
        }));
    };

    const generateInstallCommand = () => {
        let command = "bash <(curl -sSL https://get.openpanel.co/)";
        for (const option in installOptions) {
            if (installOptions[option]) {
                if (typeof installOptions[option] === "boolean") {
                    command += ` --${option}`;
                } else {
                    command += ` --${option}=${installOptions[option]}`;
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
                                    <label htmlFor={key}>{key.replace(/([A-Z])/g, ' $1').toUpperCase()}:</label>
                                    {key === "version" && (
                                        <select
                                            id={key}
                                            name={key}
                                            value={value}
                                            onChange={handleInputChange}
                                        >
                                            {availableVersions.map(version => (
                                                <option key={version} value={version}>{version}</option>
                                            ))}
                                        </select>
                                    )}
                                    {key === "screenshots" && (
                                        <select
                                            id={key}
                                            name={key}
                                            value={value}
                                            onChange={handleInputChange}
                                        >
                                            <option value="local">Local</option>
                                            <option value="http://screenshots-api.openpanel.co/screenshot">http://screenshots-api.openpanel.co/screenshot</option>
                                            <option value="custom">Custom</option>
                                        </select>
                                    )}
                                    {typeof value === "boolean" ? (
                                        <input
                                            type="checkbox"
                                            id={key}
                                            name={key}
                                            checked={value}
                                            onChange={handleInputChange}
                                        />
                                    ) : (
                                        <input
                                            type="text"
                                            id={key}
                                            name={key}
                                            value={value}
                                            onChange={handleInputChange}
                                        />
                                    )}
                                </div>
                            ))}
                        </div>
                    </div>

                    <button className="bg-blue-500 text-white px-4 py-2" onClick={generateInstallCommand}>
                        Generate Install Command
                    </button>
                </div>
                <BlogFooter />
            </div>
        </CommonLayout>
    );
};

export default Install;
