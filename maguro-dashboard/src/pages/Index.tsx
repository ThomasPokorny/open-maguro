import { useState } from "react";
import { useQuery } from "@tanstack/react-query";
import { listAgents, listSkills } from "@/lib/api";
import { AgentsView } from "@/components/AgentsView";
import { SkillsView } from "@/components/SkillsView";
import { KanbanView } from "@/components/KanbanView";
import { MaguroChatView } from "@/components/MaguroChatView";
import { ExecutionLogsPanel } from "@/components/ExecutionLogsPanel";
import { AppSidebar } from "@/components/AppSidebar";
import { Toaster } from "@/components/ui/toaster";
import logoImg from "@/assets/logo.png";

type Tab = "chat" | "agents" | "skills" | "board" | "logs";

const Index = () => {
  const [activeTab, setActiveTab] = useState<Tab>("chat");
  const [selectedTeamId, setSelectedTeamId] = useState<string | null>(null);
  const [logsOpen, setLogsOpen] = useState(false);

  const {
    data: agents = [],
    isLoading: agentsLoading,
    refetch: refetchAgents,
  } = useQuery({
    queryKey: ["agents", selectedTeamId],
    queryFn: () => listAgents(selectedTeamId ? { team_id: selectedTeamId } : undefined),
    staleTime: 30_000,
  });

  const {
    data: skills = [],
    isLoading: skillsLoading,
    refetch: refetchSkills,
  } = useQuery({
    queryKey: ["skills"],
    queryFn: listSkills,
    staleTime: 30_000,
  });

  const handleTabChange = (tab: Tab) => {
    if (tab === "logs") {
      setLogsOpen(true);
    } else {
      setActiveTab(tab);
    }
  };

  return (
      <div className="min-h-screen bg-background flex">
        {/* Sidebar */}
        <AppSidebar
            activeTab={activeTab}
            onTabChange={handleTabChange}
            selectedTeamId={selectedTeamId}
            onTeamChange={setSelectedTeamId}
        />

        {/* Main area */}
        <div className="flex-1 flex flex-col min-w-0 h-screen">
          {/* Minimal header — hidden for chat */}
          {activeTab !== "chat" && (
              <header className="border-b border-border bg-card sticky top-0 z-40 h-14 flex items-center px-6">
                <h1 className="text-sm font-semibold text-foreground">
                  {activeTab === "agents" && (selectedTeamId ? "Agents" : "All Agents")}
                  {activeTab === "board" && "Board"}
                  {activeTab === "skills" && "Skills"}
                </h1>
              </header>
          )}

          {/* Chat — always mounted to preserve history, hidden when on other tabs */}
          <div className={activeTab === "chat" ? "flex flex-col flex-1 min-h-0" : "hidden"}>
            <MaguroChatView />
          </div>

          {/* Content for other tabs */}
          {activeTab !== "chat" && (
              <main className="flex-1 max-w-5xl mx-auto w-full px-6 py-6 pb-20">
                {activeTab === "agents" && (
                    <AgentsView
                        agents={agents}
                        allSkills={skills}
                        isLoading={agentsLoading}
                        onRefetch={refetchAgents}
                    />
                )}
                {activeTab === "skills" && (
                    <SkillsView
                        skills={skills}
                        isLoading={skillsLoading}
                        onRefetch={refetchSkills}
                    />
                )}
                {activeTab === "board" && (
                    <KanbanView agents={agents} selectedTeamId={selectedTeamId} />
                )}
              </main>
          )}
        </div>

        {/* Execution Logs slide-up panel */}
        <ExecutionLogsPanel open={logsOpen} onClose={() => setLogsOpen(false)} />
        <Toaster />

        {/* Corner logo watermark */}
        <img
            src={logoImg}
            alt="OpenMaguro logo"
            className="fixed bottom-5 right-5 w-48 h-48 object-contain opacity-40 pointer-events-none select-none z-50"
        />
      </div>
  );
};

export default Index;