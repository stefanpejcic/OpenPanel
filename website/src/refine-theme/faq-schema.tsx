import React from "react";

type FaqItem = {
    question: string;
    answer: string;
};

type Props = {
    faq: FaqItem[];
};

// Renders FAQPage structured data (https://schema.org/FAQPage) so search
// engines can show these questions as rich results.
export const FaqSchema: React.FC<Props> = ({ faq }) => {
    const jsonLd = {
        "@context": "https://schema.org",
        "@type": "FAQPage",
        mainEntity: faq.map((item) => ({
            "@type": "Question",
            name: item.question,
            acceptedAnswer: {
                "@type": "Answer",
                text: item.answer,
            },
        })),
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
