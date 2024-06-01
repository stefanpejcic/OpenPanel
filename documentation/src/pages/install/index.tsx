import React, { useState, useEffect } from "react";
import Head from "@docusaurus/Head";
import { BlogFooter } from "@site/src/refine-theme/blog-footer";
import { CommonHeader } from "@site/src/refine-theme/common-header";
import { CommonLayout } from "@site/src/refine-theme/common-layout";
import clsx from "clsx";

const Install: React.FC = () => {
    const [installOptions, setInstallOptions] = useState({
        "hostname": { value: "", description: "Set the hostname." },
        "version": { value: "", description: "Set a custom OpenPanel version to be installed." },
        "skip-requirements": { value: false, description: "Skip the requirements check." },
        "skip-panel-check": { value: false, description: "Skip checking if existing panels are installed." },
        "skip-apt-update": { value: false, description: "Skip the APT update." },
        "overlay2": { value: false, description: "Enable overlay2 storage driver instead of device-mapper." },
        "skip-firewall": { value: false, description: "Skip UFW setup UFW - Only do this if you will set another Firewall manually!" },
        "skip-images": { value: false, description: "Skip installing openpanel/nginx and openpanel/apache docker images." },
        "skip-blacklists": { value: false, description: "Do not set up IP sets and blacklists." },
        "skip-ssl": { value: false, description: "Skip SSL setup." },
        "with-modsec": { value: false, description: "Enable ModSecurity for Nginx." },
        "ips": { value: "", description: "Whitelist IP addresses of OpenPanel Support Team." },
        "no-ssh": { value: false, description: "Disable port 22 and whitelist the IP address of user installing the panel." },
        "enable-ftp": { value: false, description: "Install FTP (experimental)." },
        "enable-mail": { value: false, description: "Install Mail (experimental)." },
        "post-install": { value: "", description: "Specify the post install script path." },
        "screenshots": { value: "remote", description: "Set the screenshots API URL.", options: ["local", "remote"] }, // select
        "debug": { value: false, description: "Display debug information during installation." },
        "repair": { value: false, description: "Retry and overwrite everything." },
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

    const handleInputChange = (e: React.ChangeEvent<HTMLInputElement | HTMLSelectElement>) => {
        const { name, value, type, checked } = e.target;
        setInstallOptions(prevState => ({
            ...prevState,
            [name]: { ...prevState[name], value: type === "checkbox" ? checked : value },
        }));
    };

    const generateInstallCommand = () => {
        let command = "bash <(curl -sSL https://get.openpanel.co/)";
        for (const option in installOptions) {
            if (option !== "version" || (option === "version" && installOptions[option].value !== latestVersion)) {
                if (installOptions[option].value !== "") {
                    if (option === "screenshots" && installOptions[option].value === "local") {
                        command += ` --screenshots=local`;
                    } else if (option !== "screenshots") {
                        command += ` --${option}${installOptions[option].value === true ? '' : `=${installOptions[option].value}`}`;
                    }
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
                    <p className="mb-0">Connect to your server as root via SSH and run the following command:</p>

                    <div className="mb-0">
                        <pre className="my-0" style={{ whiteSpace: "pre-wrap" }}>{generateInstallCommand()}</pre>
                    </div>

                    <div className="mb-2">
                        <h2>Advanced Install Settings:</h2>
                        <p className="mt-0">Here you can set what shall be installed and configured when installing OpenPanel:</p>
                        <ul>
                            {/* Input fields for advanced install settings */}
                            {Object.entries(installOptions).map(([key, value]) => (
                                <li key={key}>
                                    <label htmlFor={key}>{key.replace(/-/g, ' ').toUpperCase()}:</label>
                                    <span>{value.description}</span>
                                    {key === "version" ? (
                                        <input
                                            type="text"
                                            id={key}
                                            name={key}
                                            value={value.value}
                                            onChange={handleInputChange}
                                        />
                                    ) : key === "screenshots" ? ( // Modified
                                        <select // Modified
                                            id={key} // Modified
                                            name={key} // Modified
                                            value={value.value} // Modified
                                            onChange={handleInputChange} // Modified
                                        > // Modified
                                            {value.options.map(option => ( // Modified
                                                <option key={option} value={option}>{option}</option> // Modified
                                            ))} // Modified
                                        </select> // Modified
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
