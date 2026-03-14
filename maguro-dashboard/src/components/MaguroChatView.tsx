import { useState, useEffect, useRef, useCallback } from "react";
import { useMutation } from "@tanstack/react-query";
import { sendChatMessage, resetChatSession } from "@/lib/api";
import { cn } from "@/lib/utils";
import { Button } from "@/components/ui/button";
import { RefreshCw, Send } from "lucide-react";

// ── Types ─────────────────────────────────────────────────────────────────────

interface Message {
    id: string;
    role: "user" | "maguro";
    text: string;
    animating?: boolean;
}

// ── Simple Markdown Renderer ──────────────────────────────────────────────────

function renderMarkdown(text: string): React.ReactNode[] {
    return text.split("\n").map((line, i) => {
        // Process inline markdown: **bold** and `code`
        const parts: React.ReactNode[] = [];
        const regex = /(\*\*(.+?)\*\*|`([^`]+)`)/g;
        let lastIndex = 0;
        let match;

        while ((match = regex.exec(line)) !== null) {
            if (match.index > lastIndex) {
                parts.push(line.slice(lastIndex, match.index));
            }
            if (match[2]) {
                parts.push(<strong key={match.index}>{match[2]}</strong>);
            } else if (match[3]) {
                parts.push(
                    <code
                        key={match.index}
                        className="bg-secondary/60 text-primary font-mono text-xs px-1 py-0.5 rounded"
                    >
                        {match[3]}
                    </code>
                );
            }
            lastIndex = match.index + match[0].length;
        }

        if (lastIndex < line.length) {
            parts.push(line.slice(lastIndex));
        }

        return (
            <span key={i}>
        {parts.length > 0 ? parts : line}
                {i < text.split("\n").length - 1 && <br />}
      </span>
        );
    });
}

// ── Typewriter Message ────────────────────────────────────────────────────────

function TypewriterMessage({ text, onDone }: { text: string; onDone: () => void }) {
    const [displayed, setDisplayed] = useState("");
    const doneRef = useRef(false);

    useEffect(() => {
        if (doneRef.current) return;
        setDisplayed("");
        let i = 0;
        const interval = setInterval(() => {
            i++;
            setDisplayed(text.slice(0, i));
            if (i >= text.length) {
                clearInterval(interval);
                doneRef.current = true;
                onDone();
            }
        }, 15);
        return () => clearInterval(interval);
    }, [text, onDone]);

    return <>{renderMarkdown(displayed)}</>;
}

// ── Fish Loading Animation ────────────────────────────────────────────────────

function FishLoader() {
    return (
        <div className="flex items-center gap-2 px-4 py-3 rounded-2xl bg-secondary/30 self-start max-w-[80px]">
            <style>{`
        @keyframes swimLeft {
          0%   { transform: translateX(18px); opacity: 0; }
          20%  { opacity: 1; }
          80%  { opacity: 1; }
          100% { transform: translateX(-18px); opacity: 0; }
        }
        .fish-swim {
          display: inline-block;
          animation: swimLeft 1.4s ease-in-out infinite;
          font-size: 15px;
        }
        .fish-swim:nth-child(2) { animation-delay: 0.3s; }
        .fish-swim:nth-child(3) { animation-delay: 0.6s; }
      `}</style>
            <span className="fish-swim">🐟</span>
            <span className="fish-swim">🐠</span>
            <span className="fish-swim">🐡</span>
        </div>
    );
}

// ── Chat Message Bubble ───────────────────────────────────────────────────────

function ChatBubble({
                        message,
                        onAnimationDone,
                    }: {
    message: Message;
    onAnimationDone: (id: string) => void;
}) {
    const isUser = message.role === "user";

    const handleDone = useCallback(() => {
        onAnimationDone(message.id);
    }, [message.id, onAnimationDone]);

    return (
        <div className={cn("flex w-full", isUser ? "justify-end" : "justify-start")}>
            <div
                className={cn(
                    "max-w-[75%] px-4 py-2.5 rounded-2xl text-sm leading-relaxed",
                    isUser
                        ? "bg-primary text-primary-foreground rounded-br-sm"
                        : "bg-secondary/30 text-foreground rounded-bl-sm"
                )}
            >
                {message.animating && !isUser ? (
                    <TypewriterMessage text={message.text} onDone={handleDone} />
                ) : (
                    renderMarkdown(message.text)
                )}
            </div>
        </div>
    );
}

// ── MaguroChatView ─────────────────────────────────────────────────────────────

export function MaguroChatView() {
    const [messages, setMessages] = useState<Message[]>([]);
    const [input, setInput] = useState("");
    const [animatingIds, setAnimatingIds] = useState<Set<string>>(new Set());
    const bottomRef = useRef<HTMLDivElement>(null);
    const textareaRef = useRef<HTMLTextAreaElement>(null);
    const didInit = useRef(false);

    const { mutate: sendMessage, isPending } = useMutation({
        mutationFn: (text: string) => sendChatMessage(text),
        onSuccess: (data) => {
            const id = crypto.randomUUID();
            setMessages((prev) => [
                ...prev,
                { id, role: "maguro", text: data.reply, animating: true },
            ]);
            setAnimatingIds((prev) => new Set(prev).add(id));
        },
        onError: (e: Error) => {
            const id = crypto.randomUUID();
            setMessages((prev) => [
                ...prev,
                {
                    id,
                    role: "maguro",
                    text: `Something went wrong: ${e.message}`,
                    animating: true,
                },
            ]);
            setAnimatingIds((prev) => new Set(prev).add(id));
        },
    });

    const { mutate: doReset } = useMutation({
        mutationFn: resetChatSession,
        onSuccess: () => {
            const id = crypto.randomUUID();
            setMessages([
                {
                    id,
                    role: "maguro",
                    text: "Session reset. Let's start fresh 🐟",
                    animating: true,
                },
            ]);
            setAnimatingIds(new Set([id]));
        },
    });

    const handleAnimationDone = useCallback((id: string) => {
        setAnimatingIds((prev) => {
            const next = new Set(prev);
            next.delete(id);
            return next;
        });
        setMessages((prev) =>
            prev.map((m) => (m.id === id ? { ...m, animating: false } : m))
        );
    }, []);

    // Fire greeting on first mount
    useEffect(() => {
        if (didInit.current) return;
        didInit.current = true;
        sendMessage("initiate chat, greet the user!");
    }, [sendMessage]);

    // Auto-scroll to bottom on new messages
    useEffect(() => {
        bottomRef.current?.scrollIntoView({ behavior: "smooth" });
    }, [messages, isPending]);

    // Auto-resize textarea
    const handleInput = (e: React.ChangeEvent<HTMLTextAreaElement>) => {
        setInput(e.target.value);
        const el = e.target;
        el.style.height = "auto";
        el.style.height = Math.min(el.scrollHeight, 96) + "px";
    };

    const handleSend = () => {
        const text = input.trim();
        if (!text || isPending) return;
        setMessages((prev) => [
            ...prev,
            { id: crypto.randomUUID(), role: "user", text },
        ]);
        setInput("");
        if (textareaRef.current) {
            textareaRef.current.style.height = "auto";
        }
        sendMessage(text);
    };

    const handleKeyDown = (e: React.KeyboardEvent<HTMLTextAreaElement>) => {
        if (e.key === "Enter" && !e.shiftKey) {
            e.preventDefault();
            handleSend();
        }
    };

    const isAnyAnimating = animatingIds.size > 0;

    return (
        <div className="flex flex-col h-full w-full bg-background relative">
            {/* New conversation button */}
            <div className="absolute top-3 right-4 z-10">
                <Button
                    variant="ghost"
                    size="sm"
                    onClick={() => doReset()}
                    disabled={isPending || isAnyAnimating}
                    className="text-muted-foreground hover:text-foreground gap-1.5 text-xs h-7 px-2"
                >
                    <RefreshCw className="w-3 h-3" />
                    New conversation
                </Button>
            </div>

            {/* Messages area */}
            <div className="flex-1 overflow-y-auto px-6 py-6 flex flex-col gap-3">
                {messages.map((msg) => (
                    <ChatBubble
                        key={msg.id}
                        message={msg}
                        onAnimationDone={handleAnimationDone}
                    />
                ))}
                {isPending && <FishLoader />}
                <div ref={bottomRef} />
            </div>

            {/* Input area */}
            <div className="flex-shrink-0 border-t border-border bg-card px-4 py-3">
                <div className="flex items-end gap-2 max-w-3xl mx-auto">
          <textarea
              ref={textareaRef}
              value={input}
              onChange={handleInput}
              onKeyDown={handleKeyDown}
              placeholder="Message Maguro 🐟..."
              disabled={isPending}
              rows={1}
              className={cn(
                  "flex-1 resize-none rounded-xl border border-border bg-input px-3 py-2.5 text-sm text-foreground",
                  "placeholder:text-muted-foreground focus:outline-none focus:ring-1 focus:ring-ring",
                  "disabled:opacity-50 min-h-[40px] max-h-24 leading-relaxed"
              )}
              style={{ height: "40px" }}
          />
                    <Button
                        onClick={handleSend}
                        disabled={isPending || !input.trim()}
                        size="icon"
                        className="h-10 w-10 rounded-xl flex-shrink-0 bg-primary text-primary-foreground hover:bg-primary/90"
                    >
                        <Send className="w-4 h-4" />
                    </Button>
                </div>
            </div>
        </div>
    );
}