import React, { useMemo, useState } from "react";
import clsx from "clsx";
import commands from "@site/src/data/opencli-commands.json";

type Command = { name: string; desc: string; usage: string };

const HOOKS_DIR = "/etc/openpanel/openpanel/hooks";

// splits a usage line into tokens, keeping "quoted parts" and [bracketed groups] together
const tokenize = (usage: string): string[] => {
    const out: string[] = [];
    let cur = "";
    let depth = 0;
    let quoted = false;
    for (const ch of usage) {
        if (ch === '"') quoted = !quoted;
        if (!quoted && ch === "[") depth++;
        if (!quoted && ch === "]") depth--;
        if (ch === " " && depth === 0 && !quoted) {
            if (cur) out.push(cur);
            cur = "";
        } else {
            cur += ch;
        }
    }
    if (cur) out.push(cur);
    return out;
};

// positional arguments of a command, in order, with a shell variable name for each
const positionals = (cmd: Command) => {
    const tokens = tokenize(cmd.usage).slice(2);
    const args: { token: string; variable: string; optional: boolean }[] = [];
    for (const t of tokens) {
        const optional = t.startsWith("[");
        const inner = t.replace(/^\[|\]$/g, "").replace(/"/g, "").trim();
        // flags and their values only reach the hook through "$@"
        if (inner.startsWith("-")) continue;
        const name = inner
            .replace(/^<|>$/g, "")
            .split(/[|=]/)[0]
            .replace(/[^A-Za-z0-9]+/g, "_")
            .replace(/^_+|_+$/g, "")
            .toUpperCase();
        if (!name || /^\d/.test(name)) continue;
        args.push({ token: t, variable: name, optional });
    }
    // the same name twice would overwrite itself, number the repeats
    const seen: Record<string, number> = {};
    return args.map((a) => {
        seen[a.variable] = (seen[a.variable] || 0) + 1;
        return seen[a.variable] > 1
            ? { ...a, variable: `${a.variable}_${seen[a.variable]}` }
            : a;
    });
};

const fieldClass = clsx(
    "w-full",
    "rounded-lg",
    "border border-gray-300 dark:border-gray-700",
    "bg-gray-0 dark:bg-gray-900",
    "text-gray-900 dark:text-gray-100",
    "px-3 py-2",
    "text-sm",
);

export default function HookGenerator() {
    const list = commands as Command[];
    const [filter, setFilter] = useState("");
    const [name, setName] = useState("user-add");
    const [when, setWhen] = useState<"pre" | "post">("pre");
    const [body, setBody] = useState("");
    const [copied, setCopied] = useState(false);

    const shown = useMemo(
        () =>
            list.filter(
                (c) =>
                    !filter ||
                    c.name.includes(filter.toLowerCase()) ||
                    c.desc.toLowerCase().includes(filter.toLowerCase()),
            ),
        [filter],
    );
    const cmd = list.find((c) => c.name === name) || list[0];
    const args = positionals(cmd);
    const file = `${HOOKS_DIR}/${when}_${cmd.name}.sh`;
    const userBody =
        body.trim() ||
        `echo "$(date '+%F %T') ${cmd.name} $*" >> /var/log/openpanel/hooks.log`;

    const script = [
        "#!/bin/bash",
        `# ${file}`,
        `# Runs ${when === "pre" ? "before" : "after"}: ${cmd.usage}`,
        `# Gets the same arguments as the command: $1 is the first one, "$@" all of them, flags included.`,
        "",
        ...(args.length
            ? args.map(
                  (a, i) =>
                      `${a.variable}="$${i + 1 > 9 ? `{${i + 1}}` : i + 1}"${
                          a.optional ? "   # optional, can be empty" : ""
                      }`,
              )
            : ["# this command takes no positional arguments, use \"$@\" to read its flags"]),
        "",
        userBody,
        "",
    ].join("\n");

    const setup = [
        `mkdir -p ${HOOKS_DIR}`,
        `nano ${file}`,
    ].join("\n");

    const copy = async () => {
        try {
            await navigator.clipboard.writeText(script);
            setCopied(true);
            setTimeout(() => setCopied(false), 1500);
        } catch {
            setCopied(false);
        }
    };

    return (
        <div
            className={clsx(
                "not-prose",
                "my-6",
                "rounded-xl",
                "border border-gray-300 dark:border-gray-700",
                "p-4 sm:p-6",
                "flex flex-col gap-4",
            )}
        >
            <div className="grid gap-4 sm:grid-cols-[1fr_auto]">
                <label className="flex flex-col gap-1.5 text-sm font-semibold text-gray-700 dark:text-gray-300">
                    Command
                    <input
                        type="search"
                        placeholder="Filter commands, e.g. domains"
                        value={filter}
                        onChange={(e) => setFilter(e.target.value)}
                        className={fieldClass}
                    />
                    <select
                        value={name}
                        onChange={(e) => setName(e.target.value)}
                        className={fieldClass}
                        size={6}
                        aria-label="Command"
                    >
                        {shown.map((c) => (
                            <option key={c.name} value={c.name}>
                                {c.name}
                                {c.desc ? ` - ${c.desc}` : ""}
                            </option>
                        ))}
                    </select>
                </label>
                <fieldset className="flex flex-col gap-1.5 text-sm">
                    <legend className="font-semibold text-gray-700 dark:text-gray-300 mb-1.5">
                        When
                    </legend>
                    {(["pre", "post"] as const).map((w) => (
                        <label
                            key={w}
                            className="flex items-center gap-2 text-gray-700 dark:text-gray-300"
                        >
                            <input
                                type="radio"
                                name="hook-when"
                                checked={when === w}
                                onChange={() => setWhen(w)}
                            />
                            {w === "pre"
                                ? "Before the command"
                                : "After the command"}
                        </label>
                    ))}
                </fieldset>
            </div>

            <p className="text-sm text-gray-600 dark:text-gray-400 m-0">
                <code>{cmd.usage}</code>
            </p>

            <label className="flex flex-col gap-1.5 text-sm font-semibold text-gray-700 dark:text-gray-300">
                Your script
                <textarea
                    rows={5}
                    value={body}
                    onChange={(e) => setBody(e.target.value)}
                    placeholder={`echo "$(date '+%F %T') ${cmd.name} $*" >> /var/log/openpanel/hooks.log`}
                    className={clsx(fieldClass, "font-mono")}
                    spellCheck={false}
                />
            </label>

            <div className="flex flex-col gap-1.5">
                <span className="text-sm font-semibold text-gray-700 dark:text-gray-300">
                    1. Create the file on the server
                </span>
                <pre className="m-0 rounded-lg bg-gray-100 dark:bg-gray-800 p-3 text-sm overflow-auto !text-gray-900 dark:!text-gray-100 !opacity-100">
                    {setup}
                </pre>
            </div>

            <div className="flex flex-col gap-1.5">
                <div className="flex items-center justify-between">
                    <span className="text-sm font-semibold text-gray-700 dark:text-gray-300">
                        2. Paste this into it
                    </span>
                    <button
                        type="button"
                        onClick={copy}
                        className={clsx(
                            "rounded-full",
                            "px-3 py-1",
                            "text-xs font-semibold",
                            "bg-refine-blue dark:bg-refine-cyan-alt",
                            "text-gray-0 dark:text-gray-900",
                        )}
                    >
                        {copied ? "Copied" : "Copy"}
                    </button>
                </div>
                <pre
                    className="m-0 rounded-lg bg-gray-100 dark:bg-gray-800 p-3 text-sm overflow-auto !text-gray-900 dark:!text-gray-100 !opacity-100"
                    data-testid="hook-script"
                >
                    {script}
                </pre>
            </div>
        </div>
    );
}
