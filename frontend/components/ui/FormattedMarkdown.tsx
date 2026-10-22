import React from "react";

interface FormattedMarkdownProps {
  content: string;
  className?: string;
  stripRedundantTitle?: boolean;
}

/**
 * Strips raw markdown header artifacts, escaped quotes, and stray asterisks attached to quotes.
 */
export function cleanMarkdownText(raw: string, stripRedundantTitle: boolean = true): string {
  if (!raw) return "";
  let s = raw.trim();

  // Strip redundant leading headings like "**Executive Summary**", "Executive Summary:", "### Executive Summary"
  if (stripRedundantTitle) {
    s = s.replace(/^(?:#+\s*|\*{1,3}\s*)?executive\s+summary(?:\s*:\s*|\*{1,3}\s*|\s*#+)?\s*/i, "");
  }

  // Unescape backslash-escaped quotes and characters
  s = s.replace(/\\"/g, '"').replace(/\\'/g, "'").replace(/\\\\/g, "\\");

  // Clean quotes that have asterisks attached inside or outside:
  // e.g. "*quote*" or “*quote*” -> “quote”
  s = s.replace(/(["“])\*+([^*]+?)\*+(["”])/g, "$1$2$3");
  // e.g. *"quote"* -> “quote”
  s = s.replace(/\*+(["“][^"“”]+?["”])\*+/g, "$1");
  // Clean dangling "* or *"
  s = s.replace(/(["“])\*+/g, "$1").replace(/\*+(["”])/g, "$1");

  return s.trim();
}

/**
 * Cleans verbatim evidence quotes by removing leading/trailing quote characters
 * and asterisks, preventing accidental double quotes on the UI.
 */
export function cleanEvidenceQuote(raw?: string): string {
  if (!raw) return "";
  let s = raw.trim();
  // Unescape backslash quotes
  s = s.replace(/\\"/g, '"').replace(/\\'/g, "'");
  // Strip outer quotes and asterisks
  s = s.replace(/^[*"'“”\s]+|[*"'“”\s]+$/g, "");
  // Clean inner asterisks attached to quotes
  s = s.replace(/(["“])\*+([^*]+?)\*+(["”])/g, "$1$2$3");
  return s.trim();
}

/**
 * Tokenizes a single line of text into React elements (bold, italic, code, quotes, badges).
 */
function renderInlineTokens(line: string, keyPrefix: string): React.ReactNode[] {
  // Regex to match bold, italic, code, and quoted phrases
  const regex = /(\*\*[^*]+?\*\*|__[^_]+?__|`[^`]+?`|(?<![a-zA-Z0-9])\*[^*\n]+?\*(?![a-zA-Z0-9])|(?<![a-zA-Z0-9])_[^_\n]+?_(?![a-zA-Z0-9])|["“][^"“”\n]+?["”])/g;

  let lastIndex = 0;
  let match: RegExpExecArray | null;
  const nodes: React.ReactNode[] = [];
  let tokenIdx = 0;

  while ((match = regex.exec(line)) !== null) {
    if (match.index > lastIndex) {
      nodes.push(
        <React.Fragment key={`${keyPrefix}-txt-${tokenIdx++}`}>
          {line.substring(lastIndex, match.index)}
        </React.Fragment>
      );
    }

    const raw = match[0];

    if ((raw.startsWith("**") && raw.endsWith("**")) || (raw.startsWith("__") && raw.endsWith("__"))) {
      const boldText = raw.slice(2, -2).trim();
      const lower = boldText.toLowerCase();

      // Render verdict badges with appropriate curatorial tones
      if (lower === "no-hire" || lower === "no_hire" || lower === "strong_no_hire") {
        nodes.push(
          <span
            key={`${keyPrefix}-b-${tokenIdx++}`}
            className="inline-flex items-center px-1.5 py-0.5 mx-0.5 rounded-xs font-mono text-[11px] font-bold uppercase tracking-wider bg-red-50 text-red-800 border border-red-200"
          >
            {boldText.replace(/_/g, " ")}
          </span>
        );
      } else if (lower === "hire" || lower === "strong_hire") {
        nodes.push(
          <span
            key={`${keyPrefix}-b-${tokenIdx++}`}
            className="inline-flex items-center px-1.5 py-0.5 mx-0.5 rounded-xs font-mono text-[11px] font-bold uppercase tracking-wider bg-emerald-50 text-emerald-800 border border-emerald-200"
          >
            {boldText.replace(/_/g, " ")}
          </span>
        );
      } else if (/^\d+(\.\d+)?\s*\/\s*\d+$/.test(boldText)) {
        // Score like 5/10
        nodes.push(
          <span
            key={`${keyPrefix}-b-${tokenIdx++}`}
            className="inline-flex items-center px-1.5 py-0.2 mx-0.5 bg-[#f5f2eb] border border-[#d5d2c8] rounded-xs font-mono text-xs font-bold text-sumi"
          >
            {boldText}
          </span>
        );
      } else {
        nodes.push(
          <strong key={`${keyPrefix}-b-${tokenIdx++}`} className="font-semibold text-sumi">
            {boldText}
          </strong>
        );
      }
    } else if (raw.startsWith("`") && raw.endsWith("`")) {
      const codeText = raw.slice(1, -1);
      nodes.push(
        <code
          key={`${keyPrefix}-c-${tokenIdx++}`}
          className="font-mono text-xs bg-[#efece4] text-sumi px-1.5 py-0.5 rounded-xs border border-[#ddd8ce]"
        >
          {codeText}
        </code>
      );
    } else if (
      (raw.startsWith("*") && raw.endsWith("*")) ||
      (raw.startsWith("_") && raw.endsWith("_"))
    ) {
      const italicText = raw.slice(1, -1);
      nodes.push(
        <em key={`${keyPrefix}-i-${tokenIdx++}`} className="italic text-sumi font-serif">
          {italicText}
        </em>
      );
    } else if (
      (raw.startsWith('"') && raw.endsWith('"')) ||
      (raw.startsWith("“") && raw.endsWith("”"))
    ) {
      const quoteText = raw.slice(1, -1).trim();
      nodes.push(
        <span
          key={`${keyPrefix}-q-${tokenIdx++}`}
          className="italic text-sumi bg-[#f5f2eb]/80 px-1.5 py-0.5 rounded-xs border border-[#e5e2da] font-serif text-[13px] mx-0.5"
        >
          &ldquo;{quoteText}&rdquo;
        </span>
      );
    } else {
      nodes.push(
        <React.Fragment key={`${keyPrefix}-txt-${tokenIdx++}`}>
          {raw}
        </React.Fragment>
      );
    }

    lastIndex = regex.lastIndex;
  }

  if (lastIndex < line.length) {
    nodes.push(
      <React.Fragment key={`${keyPrefix}-txt-${tokenIdx++}`}>
        {line.substring(lastIndex)}
      </React.Fragment>
    );
  }

  return nodes;
}

/**
 * FormattedMarkdown Component
 * Parses and formats markdown summary blocks, paragraphs, bullet points,
 * citations, and bold metadata with high typographic rigor.
 */
export default function FormattedMarkdown({
  content,
  className = "",
  stripRedundantTitle = true,
}: FormattedMarkdownProps) {
  if (!content) return null;

  const cleaned = cleanMarkdownText(content, stripRedundantTitle);
  if (!cleaned) return null;

  // Split into paragraphs by blank lines
  const rawParagraphs = cleaned.split(/\n\s*\n/).filter((p) => p.trim().length > 0);

  return (
    <div className={`space-y-4 font-sans text-sm text-sumi leading-relaxed ${className}`}>
      {rawParagraphs.map((para, pIdx) => {
        const lines = para.split("\n").map((l) => l.trim()).filter((l) => l.length > 0);

        // Check if all or some lines are bullet points
        const isBulletList = lines.every((l) => /^[-*•]\s+/.test(l));

        if (isBulletList) {
          return (
            <ul key={`p-${pIdx}`} className="space-y-2 pl-5 list-disc text-sumi-secondary my-2">
              {lines.map((line, lIdx) => {
                const itemText = line.replace(/^[-*•]\s+/, "");
                return (
                  <li key={`li-${pIdx}-${lIdx}`} className="leading-relaxed text-sumi">
                    {renderInlineTokens(itemText, `li-${pIdx}-${lIdx}`)}
                  </li>
                );
              })}
            </ul>
          );
        }

        // Render as standard paragraph
        return (
          <p key={`p-${pIdx}`} className="leading-relaxed">
            {renderInlineTokens(para.replace(/\s*\n\s*/g, " "), `p-${pIdx}`)}
          </p>
        );
      })}
    </div>
  );
}
