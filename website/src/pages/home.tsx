import FrequencyChartBlockedDomains from "@/components/FrequencyChartBlockedDomains";
import FrequencyChartTopBlockedClients from "@/components/FrequencyChartTopBlockedClients";
import MetricsCards from "@/components/metrics-card";
import PieChartRequestType from "@/components/pie-chart";
import RequestTimeline from "@/components/request-timeline";

export default function Home() {
  return (
    <>
      <MetricsCards />
      <div className="flex w-full mb-5 mt-5 gap-4">
        <RequestTimeline />
        <PieChartRequestType />
      </div>
      <div className="flex w-full mt-5 gap-4">
        <FrequencyChartBlockedDomains />
        <FrequencyChartTopBlockedClients />
      </div>
    </>
  );
}
