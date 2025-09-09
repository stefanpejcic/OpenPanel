import clsx from "clsx";
import React, { useState } from "react";
import { useLocation } from "@docusaurus/router";
import { AnimatePresence, motion } from "framer-motion";

type Props = {
    className?: string;
};

// users can submit rating(numbers displaying as a emoji) feedback and besides that they can also submit optional text feedback
// if they submit rating feedback, we show the text feedback input
// after they submit text feedback, we show a thank you message
export const DocSurveyWidget = ({ className }: Props) => {
    const refWidget = React.useRef<HTMLDivElement>(null);
    const location = useLocation();
    // users can change their rating feedback, so we need to keep track of the survey response
    const [survey, setSurvey] = useState<DocSurveyResponse | null>(null);
    // if the user submits rating feedback, we show the text feedback input
    const [isSurveyTextVisible, setIsSurveyTextVisible] = useState(false);
    // if the user submits text feedback, we show a thank you message
    const [isFinished, setIsFinished] = useState(false);
    // the user can change their rating feedback, so we need to keep track of the selected option
    const [selectedOption, setSelectedOption] = useState<SurveyOption | null>(
        null,
    );

    const handleSurveyOptionClick = async (option: SurveyOption) => {
        setSelectedOption(option);
        setIsSurveyTextVisible(true);
        // setTimeout is needed because the text view has a transition
        // we need to scroll to the bottom after the transition is finished so that the text input is visible
        setTimeout(() => {
            refWidget.current?.scrollIntoView({
                behavior: "smooth",
            });
        }, 150);

        if (survey) {
            const data = await updateSurvey({
                surveyId: survey.id,
                body: { response: option },
            });
            if (!data) return;
            setSurvey(data);
        } else {
            const data = await createSurvey({
                body: {
                    response: option,
                    entityId: location.pathname,
                },
            });
            if (!data) return;
            setSurvey(data);
        }
    };

    const handleSurveyTextSubmit = async (text: string) => {
        if (text.trim() === "") {
            return;
        }

        const currentSurveyId = survey?.id || generateRandomSurveyId();
        
        const data = await updateSurvey({
            surveyId: currentSurveyId,
            body: { response: selectedOption, responseText: text },
        });
        if (!data) return;

        setSurvey(data);
        // when the user submits text feedback, we show a thank you message
        setIsFinished(true);

        // reset the survey after N seconds so that the user can submit another survey
        setTimeout(() => {
            setSelectedOption(null);
            setSurvey(null);
            setIsFinished(false);
            setIsSurveyTextVisible(false);
        }, 3000);
    };

    return (
        <div
            ref={refWidget}
            className={clsx(
                "w-full max-w-[416px]",
                "flex flex-col",
                "p-3",
                "bg-gray-100 dark:bg-gray-700",
                "border border-gray-300 dark:border-gray-700",
                "rounded-[28px]",
                (isSurveyTextVisible || isFinished) && "h-[286px] sm:h-[242px]",
                !isSurveyTextVisible && !isFinished && "h-[114px] sm:h-[58px]",
                "transition-all duration-200 ease-in-out",
                "overflow-hidden",
                className,
            )}
        >
            {isFinished ? (
                <AnimatePresence>
                    <SurveyFinished selectedOption={selectedOption} />
                </AnimatePresence>
            ) : (
                <>
                    <SurveyOptions
                        options={surveyOptions}
                        selectedOption={selectedOption}
                        onOptionClick={handleSurveyOptionClick}
                    />
                    {isSurveyTextVisible && (
                        <SurveyText
                            className={clsx(
                                "w-full",
                                "mt-4",
                                isSurveyTextVisible && "h-[128px] block",
                                !isSurveyTextVisible && "h-[0px] hidden",
                                "transition-all duration-200 ease-in-out",
                            )}
                            onSubmit={handleSurveyTextSubmit}
                        />
                    )}
                </>
            )}
        </div>
    );
};

const SurveyOptions = (props: {
    className?: string;
    options: {
        value: SurveyOption;
        img: string;
    }[];
    selectedOption: SurveyOption | null;
    onOptionClick: (option: SurveyOption) => void;
}) => {
    const hasSelectedOption = !!props.selectedOption;

    return (
        <div
            className={clsx(
                "w-full",
                "flex flex-col sm:flex-row",
                "items-center justify-between",
                "gap-4 sm:gap-2",
                "sm:pl-4",
                props.className,
            )}
        >
            <div
                className={clsx(
                    "dark:text-gray-100 text-gray-800",
                    "text-base",
                )}
            >
                Was this helpful?
            </div>
            <div className={clsx("flex", "items-center", "gap-3 sm:gap-1")}>
                {props.options.map(({ value, img }) => {
                    const isSelected = props.selectedOption === value;

                    return (
                        <button
                            key={value}
                            onClick={() => props.onOptionClick(value)}
                            className="p-1.5 sm:p-1"
                        >
                            <img
                                src={img}
                                alt={img.split("/").pop()}
                                loading="lazy"
                                className={clsx(
                                    "block",
                                    "flex-shrink-0",
                                    "sm:w-6 sm:h-6",
                                    "w-9 h-9",
                                    isSelected && "mix-blend-normal",
                                    isSelected && "scale-[1.33]",
                                    !isSelected && "mix-blend-luminosity",
                                    !isSelected &&
                                        hasSelectedOption &&
                                        "opacity-50",
                                    "transition-all duration-200 ease-in-out",
                                )}
                            />
                        </button>
                    );
                })}
            </div>
        </div>
    );
};

const SurveyText = (props: {
    className?: string;
    onSubmit: (text: string) => void;
}) => {
    const [text, setText] = useState("");
    return (
        <form
            className={clsx(
                "w-full",
                "h-full",
                "flex",
                "flex-col",
                props.className,
            )}
            onSubmit={(e) => {
                e.preventDefault();
                props.onSubmit(text);
            }}
        >
            <textarea
                name="survey-text"
                required
                className={clsx(
                    "w-full",
                    "h-32",
                    "p-4",
                    "text-sm",
                    "dark:placeholder-gray-500 placeholder-gray-400",
                    "dark:text-gray-500 text-gray-400",
                    "dark:bg-gray-900 bg-white",
                    "border",
                    "dark:border-gray-700",
                    "border-gray-300",
                    "rounded-lg",
                    "resize-none",
                )}
                placeholder="Your emoji tells us how you feel. If you have any additional thoughts or suggestions, we'd love to hear them!"
                value={text}
                onChange={(e) => {
                    setText(e.currentTarget.value);
                }}
            />
            <div
                className={clsx("flex", "items-center", "justify-end", "mt-2")}
            >
                <button
                    type="submit"
                    className={clsx(
                        "w-20 h-8",
                        "text-xs",
                        "text-white ",
                        "bg-gray-600",
                        "border",
                        "border-transparent",
                        "rounded-full",
                    )}
                >
                    Send
                </button>
            </div>
        </form>
    );
};

const SurveyFinished = (props: {
    className?: string;
    selectedOption: SurveyOption | null;
}) => {
    const option = surveyOptions.find(
        (option) => option.value === props.selectedOption,
    );

    return (
        <div
            className={clsx(
                "flex",
                "flex-col",
                "items-center",
                "justify-center",
                "h-full",
                "dark:text-white text-gray-800",
                props.className,
            )}
        >
            <img
                src={option?.img}
                className={clsx("w-8 h-8", "block")}
                alt="emoji"
            />
            <motion.div
                initial={{ opacity: 0 }}
                animate={{ opacity: 1, transition: { delay: 0.1 } }}
                exit={{ opacity: 0 }}
            >
                <div className={clsx("mt-6")}>Thank you!</div>
            </motion.div>
            <motion.div
                initial={{ opacity: 0 }}
                animate={{ opacity: 1, transition: { delay: 0.2 } }}
                exit={{ opacity: 0 }}
            >
                <div className={clsx("mt-1")}>
                    Your feedback has been recieved.
                </div>
            </motion.div>
        </div>
    );
};

const generateRandomSurveyId = () => {
    return Math.random().toString(36).substring(7);
};

const createSurvey = async ({ body }: { body: DocSurveyCreateDto }) => {
    const surveyId = generateRandomSurveyId();
    const response = await fetch(`${DOC_SURVEY_URL}?surveyId=${surveyId}`, {
        method: "POST",
        headers: {
            "Content-Type": "application/json",
        },
        body: JSON.stringify(body),
    });

    if (!response.ok) {
        return null;
    }

    const data: DocSurveyResponse = await response.json();
    return data;
};



const updateSurvey = async ({
    surveyId,
    body,
}: {
    surveyId?: string;
    body: DocSurveyUpdateDto;
}) => {
    const response = await fetch(`${DOC_SURVEY_URL}?surveyId=${surveyId}`, {
        method: "PATCH",
        headers: {
            "Content-Type": "application/json",
        },
        body: JSON.stringify(body),
    });

    if (!response.ok) {
        return null;
    }

    const data: DocSurveyResponse = await response.json();
    return data;
};


const surveyOptions: {
    value: SurveyOption;
    img: string;
}[] = [
    {
        value: 1,
        img: "/icons/emoji/emoji-crying-face.png",
    },
    {
        value: 2,
        img: "/icons/emoji/emoji-sad-face.png",
    },
    {
        value: 3,
        img: "/icons/emoji/emoji-neutral-face.png",
    },
    {
        value: 4,
        img: "/icons/emoji/emoji-slightly-smiling-face.png",
    },
    {
        value: 5,
        img: "/icons/emoji/emoji-star-struct-face.png",
    },
];

export type SurveyOption = 1 | 2 | 3 | 4 | 5;

export type Survey = {
    id: string;
    name: string;
    slug: string;
    options: SurveyOption[];
    source: string;
    entityType: string;
    surveyType: string;
    createdAt: string;
    updatedAt: string;
};

export type IDocSurveyContext = {
    survey: Survey;
};

export type DocSurveyCreateDto = {
    response: number;
    entityId: string;
    responseText?: string;
    metaData?: SurveyMetaData;
};

export type DocSurveyUpdateDto = {
    response: number;
    responseText?: string;
    metaData?: SurveyMetaData;
};

export type DocSurveyResponse = {
    response: number;
    entityId: string;
    survey: Survey;
    responseText?: string | null;
    metaData: SurveyMetaData;
    id: string;
    createdAt: string;
    updatedAt: string;
};

export type SurveyMetaData = Record<string, any>;

const DOC_SURVEY_URL = `https://api.openpanel.com/surveys/documentation-pages-survey.php`;
