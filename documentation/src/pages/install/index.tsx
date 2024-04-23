import React, { useState } from "react";
import Head from "@docusaurus/Head";
import { BlogFooter } from "@site/src/refine-theme/blog-footer";
import { CommonHeader } from "@site/src/refine-theme/common-header";
import { CommonLayout } from "@site/src/refine-theme/common-layout";
import clsx from "clsx";

const Install: React.FC = () => {
  const [options, setOptions] = useState({
    port: false,
    lang: false,
    hostname: false,
    email: false,
    password: false,
    apache: true, // Default to true
    phpfpm: true, // Default to true
    // Add more options as needed
  });

  const [inputValues, setInputValues] = useState({
    // Define initial input values if needed
    portValue: "",
    langValue: "",
    hostnameValue: "",
    emailValue: "",
    passwordValue: "",
  });

  const handleOptionChange = (option: string) => {
    setOptions((prevOptions) => ({
      ...prevOptions,
      [option]: !prevOptions[option],
    }));
  };

  const handleInputChange = (event: React.ChangeEvent<HTMLInputElement>) => {
    const { name, value } = event.target;
    setInputValues((prevInputValues) => ({
      ...prevInputValues,
      [name]: value,
    }));
  };

  const generateInstallCommand = () => {
    let command = "Your install command here";
    // Logic to generate install command based on selected options and input values
    // For example:
    if (options.port) {
      command += ` --port ${inputValues.portValue}`;
    }
    if (options.lang) {
      command += ` --lang ${inputValues.langValue}`;
    }
    // Add more options as needed
    return command;
  };

  return (
    <CommonLayout>
      <Head title="Install | OpenPanel">
        <html data-page="support" data-customized="true" />
      </Head>
      <div className="refine-prose">
        <CommonHeader hasSticky={true} />

        <div className="flex-1 flex flex-col pt-8 lg:pt-16 pb-32 max-w-[800px] w-full mx-auto px-2">
          <h2 className="u-text-center">Installation instructions</h2>
          <p className="u-mb10">
            Log in to your server e.g.{" "}
            <code>ssh root@your.server</code> then download the installation
            script:
          </p>
          {/* Implement CopyToClipboardInput for first command */}

          <p className="u-mb10">
            Check you are running as the <code>root</code> user, configure the
            options you want below, then run:
          </p>
          {/* Implement CopyToClipboardInput for second command */}

          <h2 className="u-text-center">Configure options</h2>
          <ul className="option-list">
            {Object.keys(options).map((option) => (
              <li
                key={option}
                className={clsx("option-item", "is-clickable", {
                  "is-active": options[option],
                })}
              >
                <div className="option-header">
                  <div className="form-check">
                    <input
                      type="checkbox"
                      className="form-check-input"
                      id={option}
                      checked={options[option]}
                      onChange={() => handleOptionChange(option)}
                    />
                    <label htmlFor={option}>{option}</label>
                  </div>
                  {/* Implement tooltip icon */}
                </div>
                {/* Input fields for each option */}
                {options[option] && (
                  <input
                    type="text"
                    name={`${option}Value`}
                    value={inputValues[`${option}Value`]}
                    onChange={handleInputChange}
                    placeholder={`Enter value for ${option}`}
                  />
                )}
              </li>
            ))}
          </ul>

          {/* Implement input fields for additional input values */}
          {/* Implement a button to generate the install command */}
          <button
            className="bg-blue-500 text-white px-4 py-2"
            onClick={() => generateInstallCommand()}
          >
            Generate Install Command
          </button>
          {/* Display the generated command */}
          <p>{generateInstallCommand()}</p>
        </div>
        <BlogFooter />
      </div>
    </CommonLayout>
  );
};

export default Install;
