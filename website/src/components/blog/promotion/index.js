/* eslint-disable react/jsx-no-target-blank */
import React from "react";

const PromotionBanner = ({ image, title, description, isDark }) => {
    let renderDescription;

    if (description === "refineNew") {
        renderDescription = (
            <div>
                {" "}
                <a href="https://openpanel.com" target="_blank">
                    refine.new
                </a>{" "}
                enables you to create React-based, headless UI enterprise
                applications within your browser that you can preview, tweak and
                download instantly.
                <br />
                <br />
                🚀 By visually combining options for your preferred ✨
                <b> React platform,</b> ✨ <b>UI framework</b>, ✨{" "}
                <b>backend connector</b>, and ✨ <b>auth provider</b>; you can
                create tailor-made architectures for your project in seconds. It
                feels like having access to thousands of project templates at
                your fingertips, allowing you to choose the one that best suits
                your needs!
                <br />
                <br />
                <br />
            </div>
        );
    } else {
        renderDescription = description ?? (
            <div>
                Meet the headless, React-based solution to build sleek{" "}
                <b>CRUD</b> applications. With refine, you can be confident that
                your codebase will always stay clean and boilerplate-free.
                <br />
                <br />
                Try{" "}
                <a href="https://github.com/refinedev/refine" target="_blank">
                    refine
                </a>{" "}
                to rapidly build your next <b>CRUD</b> project, whether {"it's"}{" "}
                an admin panel, dashboard, internal tool or storefront.
            </div>
        );
    }

    const imgBase =
        image ??
        "/img/generic_banner.png";

    const imgUrl = imgBase.startsWith("http")
        ? imgBase
        : imgBase.startsWith("/")
          ? imgBase
          : `/${imgBase}`;

    return (
        /*  <div className={`banner-container ${isDark && "dark"}`}>
            <div className="banner-header">
                {title ?? "Does your CRUD app need server state management?"}
            </div>

            {renderDescription}
            <div>
                <a
                    href={
                        description === "refineNew"
                            ? "https://openpanel.com"
                            : "https://github.com/refinedev/refine"
                    }
                    target="_blank"
                >
                    <img src={imgUrl} alt="refine blog logo" />
                </a>
            </div>
            <br />
        </div> */
        <a
            href="https://openpanel.com"
            target="_blank"
            rel="noreferrer"
        >
            <img src="/hackathon-2/hackathon_cover.png" />
        </a>
    );
};

export default PromotionBanner;
