import React, { useState, useEffect } from "react";
import Head from "@docusaurus/Head";
import { BlogFooter } from "@site/src/refine-theme/blog-footer";
import { CommonHeader } from "@site/src/refine-theme/common-header";
import { CommonLayout } from "@site/src/refine-theme/common-layout";
import clsx from "clsx";

const Install: React.FC = () => {
    const [installOptions, setInstallOptions] = useState({
        // Other options...
        "screenshots": { value: "local", description: "Set the screenshots API URL.", options: ["local", "remote"] }, // Modified
        // Other options...
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

    const handleInputChange = (e: React.ChangeEvent<HTMLInputElement | HTMLSelectElement>) => { // Modified
        const { name, value, type, checked } = e.target;
        setInstallOptions(prevState => ({
            ...prevState,
            [name]: { ...prevState[name], value: type === "checkbox" ? checked : value }, // Modified
        }));
    };

    const generateInstallCommand = () => {
        let command = "bash <(curl -sSL https://get.openpanel.co/)";
        for (const option in installOptions) {
            if (option !== "version" || (option === "version" && installOptions[option].value !== latestVersion)) {
                if (installOptions[option].value) {
                    if (option === "screenshots" && installOptions[option].value === "local") { // Modified
                        command += ` --screenshots=local`; // Modified
                    } else if (option !== "screenshots") {
                        command += ` --${option.replace(/-/g, '_')}=${installOptions[option].value}`;
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
                    <p>Connect to your server as root via SSH and run the following command:</p>

                    <div className="mb-2">
                        <pre style={{ whiteSpace: "pre-wrap" }}>{generateInstallCommand()}</pre>
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
                                            value={latestVersion}
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
