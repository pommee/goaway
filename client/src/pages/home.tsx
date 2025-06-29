import FrequencyChartBlockedDomains from "@/app/home/FrequencyChartBlockedDomains";
import FrequencyChartTopBlockedClients from "@/app/home/FrequencyChartTopBlockedClients";
import MetricsCards from "@/app/home/metrics-card";
import PieChartRequestType from "@/app/home/request-types";
import RequestTimeline from "@/app/home/request-timeline";
import ResponseSizeTimeline from "@/app/home/ResponseSizeTimeline";

export function Home() {
  return (
    <>
      <MetricsCards />
      <div className="flex w-full mb-5 mt-5 gap-4 flex-col sm:flex-row">
        <RequestTimeline />
        <PieChartRequestType />
      </div>
      <div className="flex w-full mb-5 mt-5 gap-4 flex-col md:flex-row">
        <div className="w-full md:w-1/2 h-[250px]">
          <FrequencyChartBlockedDomains />
        </div>
        <div className="w-full md:w-1/2 h-[250px]">
          <FrequencyChartTopBlockedClients />
        </div>
      </div>
      <ResponseSizeTimeline />
    </>
  );
}
