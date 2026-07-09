import React from "react";

type Offer = {
    price: string;
    priceCurrency: string;
    priceUnit: string;
};

type Props = {
    name: string;
    description: string;
    url: string;
    applicationCategory?: string;
    offers?: Offer;
};

// Renders SoftwareApplication structured data (https://schema.org/SoftwareApplication)
// so search engines understand this is a product page, not just marketing copy.
export const SoftwareApplicationSchema: React.FC<Props> = ({
    name,
    description,
    url,
    applicationCategory = "WebApplication",
    offers,
}) => {
    const jsonLd = {
        "@context": "https://schema.org",
        "@type": "SoftwareApplication",
        name,
        description,
        url,
        applicationCategory,
        operatingSystem: "Linux",
        ...(offers && {
            offers: {
                "@type": "Offer",
                price: offers.price,
                priceCurrency: offers.priceCurrency,
                priceSpecification: {
                    "@type": "UnitPriceSpecification",
                    price: offers.price,
                    priceCurrency: offers.priceCurrency,
                    unitText: offers.priceUnit,
                },
            },
        }),
    };

    return (
        <script
            type="application/ld+json"
            dangerouslySetInnerHTML={{
                __html: JSON.stringify(jsonLd).replace(/</g, "\\u003c"),
            }}
        />
    );
};
