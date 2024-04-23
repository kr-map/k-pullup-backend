"use client";

import Heading from "@/components/atom/Heading";
import { Input } from "@/components/ui/input";
import useAddressData from "@/hooks/common/useAddressData";
import useInput from "@/hooks/common/useInput";
import useChatIdStore from "@/store/useChatIdStore";
import getRegion from "@/utils/getRegion";
import { Fragment, useEffect, useRef, useState } from "react";
// TODO: 맵 옮길때마다 변하는거 방지

export interface ChatMessage {
  uid: string;
  message: string;
  userId: string;
  userNickname: string;
  roomID: string;
  timestamp: number;
  isOwner: boolean;
}

export interface Chatdata {
  msg: string;
  name: string;
  isOwner: boolean;
  mid: string;
  userid: string;
}

const ChatClient = () => {
  const cidState = useChatIdStore();

  const chatValue = useInput("");

  const { address, isError } = useAddressData();

  const ws = useRef<WebSocket | null>(null);
  const chatBox = useRef<HTMLDivElement>(null);
  const inputRef = useRef<HTMLInputElement>(null);

  const [connection, setConnection] = useState(false);

  const [messages, setMessages] = useState<Chatdata[]>([]);
  const [connectionMsg, setConnectionMsg] = useState("");

  const [roomTitle, setRoomTitle] = useState("");

  const [isChatError, setIsChatError] = useState(false);

  useEffect(() => {
    if (!inputRef.current) return;
    inputRef.current.focus();
  }, [inputRef]);

  useEffect(() => {
    const code = getRegion(address?.depth1 as string).getCode();
    if (!address || isError || code === "") {
      setConnectionMsg("채팅 서비스를 지원하지 않는 지역입니다!");
      return;
    }

    ws.current = new WebSocket(
      `wss://api.k-pullup.com/ws/${code}?request-id=${cidState.cid}`
    );

    ws.current.onopen = () => {
      setConnection(true);
      setConnectionMsg(
        "비속어 사용에 주의해주세요. 이후 서비스 사용이 제한될 수 있습니다!"
      );
    };

    ws.current.onmessage = async (event) => {
      const data: ChatMessage = JSON.parse(event.data);
      if (data.userNickname === "chulbong-kr") {
        const titleArr = data.message.split(" ");

        titleArr[0] = getRegion(data.roomID).getTitle();
        console.log(titleArr);

        setRoomTitle(titleArr.join(" "));
      }

      setMessages((prevMessages) => [
        ...prevMessages,
        {
          msg: data.message,
          name: data.userNickname,
          isOwner: data.isOwner,
          mid: data.uid,
          userid: data.userId,
        },
      ]);
    };

    ws.current.onerror = (error) => {
      setConnectionMsg(
        "채팅방에 참여 중 에러가 발생하였습니다. 잠시 후 다시 시도해 주세요!"
      );
      console.error("연결 에러:", error);
      setIsChatError(true);
    };

    ws.current.onclose = () => {
      setIsChatError(true);
      console.log("연결 종료");
    };

    return () => {
      ws.current?.close();
    };
  }, [address]);

  useEffect(() => {
    if (!ws) return;
    const pingInterval = setInterval(() => {
      ws.current?.send(JSON.stringify({ type: "ping" }));
    }, 30000);

    return () => {
      clearInterval(pingInterval);
    };
  }, []);

  useEffect(() => {
    const scrollBox = chatBox.current;

    if (scrollBox) {
      scrollBox.scrollTop = scrollBox.scrollHeight;
    }
  }, [messages]);

  const handleChat = () => {
    if (chatValue.value === "") return;
    ws.current?.send(chatValue.value);
    chatValue.resetValue();
    inputRef.current?.focus();
  };

  const handleKeyPress = (event: React.KeyboardEvent<HTMLInputElement>) => {
    if (event.key === "Enter") {
      handleChat();
    }
  };

  if (isChatError) {
    return (
      <div>
        <Heading title={`서울 채팅방`} subTitle="1명 접속 중" />
        <div className="text-red text-center">
          채팅을 불러오는데 실패하였습니다. <br /> 잠시 후 다시 시도해 주세요.
        </div>
      </div>
    );
  }

  return (
    <div className="flex flex-col h-[calc(100%-96px)]">
      <Heading title={`서울 채팅방`} subTitle="1명 접속 중" />
      <div className="text-center text-sm text-grey-dark">{roomTitle}</div>
      <div
        className="grow w-full flex flex-col justify-between px-3"
        ref={chatBox}
      >
        <div className="text-center text-red">
          {connection ? connectionMsg : "채팅방 접속중..."}
        </div>
        <div>
          {messages.map((message) => {
            if (message.name === "chulbong-kr") return;
            if (message.msg?.includes("님이 입장하셨습니다.")) {
              return (
                <div
                  key={message.mid}
                  className="truncate px-5 py-2 text-center text-sm text-grey-dark"
                >
                  {message.name}님이 참여하였습니다.
                </div>
              );
            }
            if (message.msg?.includes("님이 퇴장하셨습니다.")) {
              return (
                <div
                  key={message.mid}
                  className="truncate px-5 py-2 text-center text-sm text-grey-dark"
                >
                  {message.name}님이 나가셨습니다.
                </div>
              );
            }
            return (
              <Fragment key={message.mid}>
                {message.userid === cidState.cid ? (
                  <div className="flex flex-col items-end w-full">
                    <div className="max-w-1/2 p-1 rounded-lg bg-slate-700 shadow-sm">
                      {message.msg}
                    </div>
                    <div className="text-xs text-grey-dark">{message.name}</div>
                  </div>
                ) : (
                  <div className="flex flex-col items-start w-full">
                    <div className="max-w-1/2 p-1 rounded-lg bg-slate-600 shadow-sm">
                      {message.msg}
                    </div>
                    <div className="text-xs text-grey-dark">{message.name}</div>
                  </div>
                )}
              </Fragment>
            );
          })}
        </div>
      </div>
      <div className="flex items-center justify-center w-full h-14 px-3">
        <Input
          type="text"
          ref={inputRef}
          maxLength={40}
          disabled={!connection}
          name="reveiw-content"
          value={chatValue.value}
          onChange={chatValue.handleChange}
          onKeyDown={handleKeyPress}
          className="bg-black-light-2 text-base"
        />
      </div>
    </div>
  );
};

export default ChatClient;

{
  /* <div className="truncate px-5 py-2 text-center text-sm text-grey-dark">
            asdasd님이 참여하였습니다.
          </div>
          <div className="flex flex-col items-start w-full">
            <div className="max-w-1/2 p-1 rounded-lg bg-slate-600 shadow-sm">
              ㅁㄴㅇㅁㄴㅇㅁㄴㅇㅁㄴ
            </div>
            <div className="text-xs text-grey-dark">adkwnq</div>
          </div>
          <div className="flex flex-col items-end w-full">
            <div className="max-w-1/2 p-1 rounded-lg bg-slate-700 shadow-sm">
              asdasdasdasd
            </div>
            <div className="text-xs text-grey-dark">zxczxcas</div>
          </div> */
}