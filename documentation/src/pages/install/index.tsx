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
        "--skip-requirements": false,
        "--skip-panel-check": false,
        "--skip-apt-update": false,
        overlay2: false,
        "--skip-firewall": false,
        "--skip-images": false,
        "--skip-blacklists": false,
        "--skip-ssl": false,
        "--with-modsec": false,
        ips: "",
        "--no-ssh": false,
        "--enable-ftp": false,
        "--enable-mail": false,
        "--post-install": "",
        "--screenshots": "local",
        "--debug": false,
        "--repair": false,
    });

    useEffect(() => {
        fetch("https://update.openpanel.co/")
            .then(response => response.text())
            .then(data => {
                // Extract the version number from the plain text response
                const latestVersion = data.trim();
                setInstallOptions(prevState => ({
                    ...prevState,
                    version: latestVersion,
                }));
            })
            .catch(error => console.error("Error fetching available version:", error));
    }, []);

    const handleInputChange = (e: React.ChangeEvent<HTMLInputElement>) => {
        const { name, value, type, checked } = e.target;
        setInstallOptions(prevState => ({
            ...prevState,
            [name]: type === "checkbox" ? checked : value,
        }));
    };

    const optionDescriptions = {
        hostname: "Set the hostname.",
        version: "Set a custom OpenPanel version to be installed.",
        "--skip-requirements": "Skip the requirements check.",
        "--skip-panel-check": "Skip checking if existing panels are installed.",
        "--skip-apt-update": "Skip the APT update.",
        overlay2: "Enable overlay2 storage driver instead of device-mapper.",
        "--skip-firewall": "Skip UFW setup UFW - Only do this if you will set another Firewall manually!",
        "--skip-images": "Skip installing openpanel/nginx and openpanel/apache docker images.",
        "--skip-blacklists": "Do not set up IP sets and blacklists.",
        "--skip-ssl": "Skip SSL setup.",
        "--with-modsec": "Enable ModSecurity for Nginx.",
        ips: "Whitelist IP addresses of OpenPanel Support Team.",
        "--no-ssh": "Disable port 22 and whitelist the IP address of user installing the panel.",
        "--enable-ftp": "Install FTP (experimental).",
        "--enable-mail": "Install Mail (experimental).",
        "--post-install": "Specify the post install script path.",
        "--screenshots": "Set the screenshots API URL.",
        "--debug": "Display debug information during installation.",
        "--repair": "Retry and overwrite everything.",
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
                        <pre>
                            {`bash <(curl -sSL https://get.openpanel.co/) ${
                                Object.entries(installOptions)
                                    .filter(([_, value]) => value !== "")
                                    .map(([key, value]) =>
                                        typeof value === "boolean" ? `--${key}` : `--${key}=${value}`
                                    )
                                    .join(" ")}
                            `}
                        </pre>
                    </div>

                    <div className="mb-8">
                        <h2>Advanced Install Settings:</h2>
                        <div>
                            {Object.entries(installOptions).map(([key, value]) => (
                                <div key={key} className="mb-4">
                                    <p>
                                        <strong>{key.replace(/-/g, ' ').toUpperCase()}:</strong>{" "}
                                        {optionDescriptions[key]}
                                    </p>
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
                </div>
                <BlogFooter />
            </div>
        </CommonLayout>
    );
};

export default Install;
