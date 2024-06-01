import React, { useState } from "react";
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
        screenshots: "",
        debug: false,
        repair: false,
    });

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
                                    <input
                                        type={typeof value === "boolean" ? "checkbox" : "text"}
                                        id={key}
                                        name={key}
                                        checked={typeof value === "boolean" ? value : undefined}
                                        value={typeof value === "boolean" ? undefined : value}
                                        onChange={handleInputChange}
                                    />
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
