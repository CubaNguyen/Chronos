import React, { useState } from 'react';
import { Sidebar } from './components/Sidebar';
import { Dashboard } from './components/Dashboard';
import { AIChat } from './components/AIChat';
import { SemanticSearch } from './components/SemanticSearch';
import { KnowledgeGraph } from './components/KnowledgeGraph';
import { SystemHealth } from './components/SystemHealth';

export default function App() {
  const [activeTab, setActiveTab] = useState('dashboard');
  const [activeRepo, setActiveRepo] = useState('chronos-self');

  return (
    <div className="bg-[#090d16] text-slate-100 flex h-screen overflow-hidden font-sans antialiased">
      <Sidebar
        activeTab={activeTab}
        setActiveTab={setActiveTab}
        activeRepo={activeRepo}
        setActiveRepo={setActiveRepo}
      />
      <main className="flex-1 flex flex-col overflow-hidden relative">
        {activeTab === 'dashboard' && <Dashboard onNavigateToChat={() => setActiveTab('chat')} />}
        {activeTab === 'chat' && <AIChat activeRepo={activeRepo} />}
        {activeTab === 'search' && <SemanticSearch activeRepo={activeRepo} />}
        {activeTab === 'graph' && <KnowledgeGraph activeRepo={activeRepo} />}
        {activeTab === 'health' && <SystemHealth />}
      </main>
    </div>
  );
}

