import { Analyzer } from "@/components/analyzer";

export default function Home() {
  return (
    <main className="mx-auto max-w-[1180px] px-4 pt-6 pb-16">
      <header>
        <h1 className="text-[26px] font-bold">Guild Logs</h1>
        <p className="mt-1 mb-5 text-fg-2">
          Warcraft Logs report analysis: avoidable deaths, avoidable damage, activity and defensives.
        </p>
      </header>
      <Analyzer />
    </main>
  );
}
