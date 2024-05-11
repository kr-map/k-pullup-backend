"use client";

import { type ReportsRes } from "@/api/report/getMyReports";
import useReportsData from "@/hooks/query/report/useReportsData";
import { QueryObserverRefetchErrorResult } from "@tanstack/react-query";
import { isAxiosError } from "axios";
import MarkerReportList from "./_components/MarkerReportList";

interface Props {
  type?: "me" | "formarker" | "all";
  markerId?: number;
}

const ReportClient = ({ type = "me", markerId }: Props) => {
  const {
    data: myReports,
    error,
    isError,
  } = useReportsData({ type, markerId }) as QueryObserverRefetchErrorResult<
    ReportsRes[],
    Error
  >;

  if (isError) {
    if (isAxiosError(error)) {
      if (error.response?.status === 404) {
        return (
          <div className="text-center">정보 수정 제안한 위치가 없습니다.</div>
        );
      } else {
        return <div className="text-center">잠시 후 다시 시도해 주세요.</div>;
      }
    } else {
      return <div className="text-center">잠시 후 다시 시도해 주세요.</div>;
    }
  }
  // TODO: 이미지 개수 이상 오류
  console.log(myReports);

  return (
    <div>
      {myReports?.map((report) => {
        return (
          <div key={report.reportId} className="mb-4">
            <MarkerReportList
              markerId={report.markerId}
              lat={report.latitude}
              lng={report.longitude}
              desc={report.description}
              img={report.photoUrls}
              status={report.status}
            />
          </div>
        );
      })}
    </div>
  );
};

export default ReportClient;