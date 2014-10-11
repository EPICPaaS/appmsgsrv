package com.ks.gopush.cli;

import java.util.ArrayList;
import java.util.HashMap;
import java.util.List;
import java.util.Map;

import org.json.JSONArray;
import org.json.JSONException;
import org.json.JSONObject;

import com.ks.gopush.cli.model.request.BaseRequest;
import com.ks.gopush.cli.model.request.MsgRequest;
import com.ks.gopush.cli.model.response.BaseResponse;
import com.ks.gopush.cli.model.response.MsgResponse;
import com.ks.gopush.cli.utils.Constant;
import com.ks.gopush.cli.utils.JsonMapper;

public class JsonUtilTest {
	
	public static void main(String[] args) throws JSONException {
		//test1();
			test2();
			
			test3();
	}

	private static void test3() {
		// TODO Auto-generated method stub
		MsgResponse msgresp = new MsgResponse();
		BaseResponse baseResponse = new BaseResponse();
		baseResponse.setRet(Constant.OK);
		baseResponse.setErrMsg("ddd");
		
		msgresp.setBaseResponse(baseResponse);
		System.out.println(JsonMapper.buildNormalMapper().toJson(msgresp));
	}

	private static void test2() {
		MsgRequest msgr = new MsgRequest();
		msgr.setContent("ssssssssssssssss");
		msgr.setMsgType("1");
		
		BaseRequest	tmp = new BaseRequest();
		tmp.setToken("dddddddddddddddddddddddddddddd");
		msgr.setBaseRequest(tmp);
		
		String[] toUserNames = {"22622391649384526@user","22622391649384527@user"};
		msgr.setToUserNames(toUserNames);
		
		Map<String,String> objectContent = new HashMap<String, String>();
		objectContent.put("appId", "2222");
		objectContent.put("appSendId", "111");
		msgr.setObjectContent(objectContent);
		
		System.out.println(JsonMapper.buildNormalMapper().toJson(msgr));
	}

	private static void test1() throws JSONException {
		// TODO Auto-generated method stub
		//声明一个Hash对象并添加数据
		Map params =  new HashMap();

		params.put("username", "1");
		params.put("user_json", "dd");


		System.out.println(JsonMapper.buildNormalMapper().toJson(params));
	}

}
